package module

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"

	cosmosMath "cosmossdk.io/math"
	"github.com/allora-network/allora-chain/x/emissions/keeper"
	"github.com/allora-network/allora-chain/x/emissions/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// A structure to hold the original value and a random tiebreaker
type SortableItem[T any] struct {
	Value      T
	Weight     uint64
	Tiebreaker uint32
}

type RequestId = string
type BlockHeight = int64
type TopicId = uint64

type PriceAndReturn = struct {
	Price  cosmosMath.Uint
	Return cosmosMath.Uint
}

type Demand struct {
	Requests      []types.InferenceRequest
	FeesGenerated cosmosMath.Uint
}

type BuildBestPrice func(requestsForGivenTopic []types.InferenceRequest) (BestPriceData, error)

// Sorts the given slice of topics in descending order according to their corresponding return, using randomness as tiebreaker
// e.g. ([]uint64{1, 2, 3}, map[uint64]uint64{1: 2, 2: 2, 3: 3}, 0) -> [3, 1, 2] or [3, 2, 1]
func SortTopicsByReturnDescWithRandomTiebreaker(valsToSort []TopicId, weights map[TopicId]PriceAndReturn, randSeed BlockHeight) []TopicId {
	// Convert the slice of Ts to a slice of SortableItems, each with a random tiebreaker
	r := rand.New(rand.NewSource(randSeed))
	items := make([]SortableItem[TopicId], len(valsToSort))
	for i, topicId := range valsToSort {
		items[i] = SortableItem[TopicId]{topicId, weights[topicId].Price.Uint64(), r.Uint32()}
	}

	// Sort the slice of SortableItems
	// If the values are equal, the tiebreaker will decide their order
	sort.Slice(items, func(i, j int) bool {
		if items[i].Value == items[j].Value {
			return items[i].Tiebreaker > items[j].Tiebreaker
		}
		return items[i].Weight > items[j].Weight
	})

	// Extract and print the sorted values to demonstrate the sorting
	sortedValues := make([]TopicId, len(valsToSort))
	for i, item := range items {
		sortedValues[i] = item.Value
	}
	return sortedValues
}

// Check that a request:
//
//	should be checked in this timestep
//	AND did not expire
//	AND would have enough funds to cover the potential next price
//	AND admits and acceptable max price per inference
func IsValidAtPrice(
	ctx sdk.Context,
	k keeper.Keeper,
	req types.InferenceRequest,
	price cosmosMath.Uint,
	currentBlock BlockHeight,
) (bool, error) {
	reqUnmetDemand, err := k.GetRequestDemand(ctx, req.Id)
	if err != nil {
		fmt.Println("Error getting request demand: ", err)
		return false, err
	}
	res :=
		req.BlockLastChecked+req.Cadence <= currentBlock &&
			req.BlockValidUntil > currentBlock &&
			reqUnmetDemand.GTE(price) &&
			req.MaxPricePerInference.GTE(price)
	return res, nil
}

// Inactivates topics with unment demand lower than minTopicUnmetDemand
func InactivateLowDemandTopics(
	ctx context.Context,
	k keeper.Keeper,
	minTopicUnmetDemand cosmosMath.Uint,
	limit uint64,
	maxLimit uint64,
) error {
	offset := uint64(0)
	key := make([]byte, 0)
	i := uint64(0)

	for {
		pageRequest := &types.SimpleCursorPaginationRequest{Limit: limit, Key: key}
		topicsActive, pageResponse, err := k.GetActiveTopics(ctx, pageRequest)
		if err != nil {
			fmt.Println("Error getting active topics: ", err)
			return err
		}

		for _, topicId := range topicsActive {
			topicUnmetDemand, err := k.GetTopicUnmetDemand(ctx, topicId)
			if err != nil {
				fmt.Println("Error getting unmet demand: ", err)
				return err
			}
			if topicUnmetDemand.LT(minTopicUnmetDemand) {
				fmt.Printf("Inactivating topic due to no demand: %v\n", topicId)
				k.InactivateTopic(ctx, topicId)
			}
		}

		// if pageResponse.NextKey is empty then we have reached the end of the list
		if len(pageResponse.NextKey) == 0 || i > maxLimit {
			break
		}

		key = pageResponse.NextKey
		offset += limit
		i++
	}
	return nil
}

// True if the request has expired and successfully popped and depleted of its unmet demand, false otherwise
func PopAndDepleteExpiredRequests(ctx sdk.Context, k keeper.Keeper, request types.InferenceRequest) (bool, error) {
	// Check if request has expired
	if request.BlockValidUntil > ctx.BlockHeight() {
		return false, nil
	}

	requestUnmetDemand, err := k.GetRequestDemand(ctx, request.Id)
	if err != nil {
		fmt.Println("Error getting request demand: ", err)
		return false, err
	}

	// Deplete the request -- Forward all remaining funds to the ecosystem bucket
	err = k.AddTopicFeeRevenue(ctx, request.TopicId, requestUnmetDemand)
	if err != nil {
		fmt.Println("Error adding topic fee revenue: ", err)
		return false, err
	}

	// Pop the request from the mempool
	err = k.RemoveFromMempool(ctx, request.Id)
	if err != nil {
		fmt.Println("Error removing request from mempool")
		return false, err
	}

	return true, nil
}

func FilterAndResolveExpiredRequests(ctx sdk.Context, k keeper.Keeper, requests []types.InferenceRequest) ([]types.InferenceRequest, error) {
	var validRequests []types.InferenceRequest
	for _, req := range requests {
		expired, err := PopAndDepleteExpiredRequests(ctx, k, req)
		if err != nil {
			fmt.Println("Error popping and depleting expired requests: ", err)
			return nil, err
		}
		if !expired {
			validRequests = append(validRequests, req)
		}
	}
	return validRequests, nil
}

type BestPriceData struct {
	bestPrice     cosmosMath.Uint
	maxFees       cosmosMath.Uint
	validRequests []types.InferenceRequest
}

// Closure to build an online best price finder, which iteratively finds the best price for a given topic
// when fit iterative lists of inference requests for that topic
func BuildOnlineBestPriceFinder(
	ctx sdk.Context,
	k keeper.Keeper,
	currentBlock BlockHeight,
) BuildBestPrice {
	// Initialize a map of request price to map of valid requests
	// map must be of type string, because the complex type of a Uint
	// will not work for the map, equality test tests the pointer not the value
	demandCurve := make(map[string]Demand)

	bestPriceData := BestPriceData{
		bestPrice:     cosmosMath.ZeroUint(),
		maxFees:       cosmosMath.ZeroUint(),
		validRequests: make([]types.InferenceRequest, 0),
	}

	return func(requestsForGivenTopic []types.InferenceRequest) (BestPriceData, error) {
		// Loop through inference requests and then loop again (nested) checking validity of all other inferences at the first inference's max price
		for _, req := range requestsForGivenTopic {
			// Check validity of current request at its own price
			isValidAtPrice, err := IsValidAtPrice(ctx, k, req, req.MaxPricePerInference, currentBlock)
			if err != nil {
				fmt.Println("Error checking if request is valid at price: ", err)
				return BestPriceData{}, err
			}
			//fmt.Println("Req id ", req.TopicId, " is valid at price ", req.MaxPricePerInference, " : ", isValidAtPrice)
			if isValidAtPrice {
				price := req.MaxPricePerInference
				priceStr := price.String()
				_, exists := demandCurve[priceStr]
				if exists {
					// if the demand curve has already computed the list of buyers
					// at this price level before, we dont need to do it again
					continue
				}
				demandCurve[priceStr] = Demand{
					Requests:      make([]types.InferenceRequest, 0),
					FeesGenerated: cosmosMath.ZeroUint()}

				for _, req2 := range requestsForGivenTopic {
					isValidAtPrice, err := IsValidAtPrice(ctx, k, req2, price, currentBlock)
					if err != nil {
						fmt.Println("Error checking if request is valid at price: ", err)
						return BestPriceData{}, err
					}
					if isValidAtPrice {
						newFeesGenerated := demandCurve[priceStr].FeesGenerated.Add(price)
						newRequests := append(demandCurve[priceStr].Requests, req2)
						demandCurve[priceStr] = Demand{
							Requests:      newRequests,
							FeesGenerated: newFeesGenerated}
						if newFeesGenerated.GT(bestPriceData.maxFees) {
							bestPriceData.maxFees = newFeesGenerated
							bestPriceData.bestPrice = price
						}
					}
				}
			}
		}

		if !bestPriceData.bestPrice.IsZero() {
			bestPriceData.validRequests = demandCurve[bestPriceData.bestPrice.String()].Requests
		}

		return bestPriceData, nil
	}
}

func BuildBestPricePerTopicId(
	ctx sdk.Context,
	k keeper.Keeper,
	requestPageLimit uint64,
	maxRequestPages uint64,
	topicId uint64,
	buildFunc BuildBestPrice,
) (BestPriceData, error) {

	bestPriceData := BestPriceData{
		bestPrice:     cosmosMath.ZeroUint(),
		maxFees:       cosmosMath.ZeroUint(),
		validRequests: make([]types.InferenceRequest, 0),
	}
	requestPageKey := make([]byte, 0)
	j := uint64(0)
	for {
		requestPageRequest := &types.SimpleCursorPaginationRequest{Limit: requestPageLimit, Key: requestPageKey}
		inferenceRequests, requestPageResponse, err := k.GetMempoolInferenceRequestsForTopic(ctx, topicId, requestPageRequest)
		if err != nil {
			fmt.Println("Error getting mempool inference requests: ", err)
			return BestPriceData{}, err
		}
		// Filter out expired requests and deplete them
		inferenceRequests, err = FilterAndResolveExpiredRequests(ctx, k, inferenceRequests)
		if err != nil {
			fmt.Println("Error filtering and resolving expired requests: ", err)
			return BestPriceData{}, err
		}
		bestData, err := buildFunc(inferenceRequests)
		if err != nil {
			fmt.Println("Error getting requests that maximize fees: ", err)
			return BestPriceData{}, err
		}
		// Update bestPriceData from last bundle
		if bestData.maxFees.GTE(bestPriceData.maxFees) {
			bestPriceData = bestData
		}
		if len(requestPageResponse.NextKey) == 0 || j > maxRequestPages {
			break
		}
		requestPageKey = requestPageResponse.NextKey
		j++
	}

	return bestPriceData, nil
}

// The price of inference for a topic is determined by the price that maximizes the demand drawn from valid requests.
// Which topics get processed (inference solicitation and weight-adjustment) is based on ordering topics by their return
// at their optimal prices and then skimming the top.
func ChurnRequestsGetActiveTopicsAndDemand(
	ctx sdk.Context,
	k keeper.Keeper,
	currentBlock BlockHeight,
	topicPageLimit uint64,
	maxTopicPages uint64,
	requestPageLimit uint64,
	maxRequestPages uint64,
) ([]*types.Topic, cosmosMath.Uint, error) {
	minTopicDemand, err := k.GetParamsMinTopicUnmetDemand(ctx)
	if err != nil {
		fmt.Println("Error getting min topic unmet demand: ", err)
		return nil, cosmosMath.Uint{}, err
	}
	// Need to do this separate from loop below to avoid consequences from deleting things mid-iteration
	err = InactivateLowDemandTopics(ctx, k, minTopicDemand, topicPageLimit, maxTopicPages)
	if err != nil {
		return nil, cosmosMath.Uint{}, err
	}

	topicsActiveWithDemand := make([]TopicId, 0)
	topicBestPrices := make(map[TopicId]PriceAndReturn)
	requestsToDrawDemandFrom := make(map[TopicId][]types.InferenceRequest, 0)

	topicPageKey := make([]byte, 0)
	i := uint64(0)

	for {
		topicPageRequest := &types.SimpleCursorPaginationRequest{Limit: topicPageLimit, Key: topicPageKey}
		topicsActive, topicPageResponse, err := k.GetActiveTopics(ctx, topicPageRequest)
		if err != nil {
			fmt.Println("Error getting active topics: ", err)
			return nil, cosmosMath.Uint{}, err
		}
		priceFinder := BuildOnlineBestPriceFinder(ctx, k, currentBlock)
		for _, topicId := range topicsActive {

			bestPriceData, err := BuildBestPricePerTopicId(
				ctx, k, requestPageLimit, maxRequestPages, topicId, priceFinder)
			if err != nil {
				fmt.Println("Error getting requests that maximize fees: ", err)
				return nil, cosmosMath.Uint{}, err
			}
			fmt.Println("Topic: ", topicId, " Price of max return: ", bestPriceData.bestPrice, " Max return: ", bestPriceData.maxFees, " Requests to use: ", len(bestPriceData.validRequests))
			topicsActiveWithDemand = append(topicsActiveWithDemand, topicId)
			topicBestPrices[topicId] = PriceAndReturn{bestPriceData.bestPrice, bestPriceData.maxFees}
			requestsToDrawDemandFrom[topicId] = bestPriceData.validRequests
		}
		// if pageResponse.NextKey is empty then we have reached the end of the list
		if topicsActive == nil || i > maxTopicPages {
			break
		}

		topicPageKey = topicPageResponse.NextKey
		i++
	}

	// Sort topics by topicBestPrices
	sortedTopics := SortTopicsByReturnDescWithRandomTiebreaker(topicsActiveWithDemand, topicBestPrices, currentBlock)

	maxTopicsPerBlock, err := k.GetParamsMaxTopicsPerBlock(ctx)
	if err != nil {
		fmt.Println("Error getting max topics per block: ", err)
		return nil, cosmosMath.Uint{}, err
	}
	// Take top keeper.MAX_TOPICS_PER_BLOCK number of topics with the highest demand
	cutoff := uint(math.Min(float64(len(sortedTopics)), float64(maxTopicsPerBlock)))

	topTopicsByReturn := sortedTopics[:cutoff]
	// Reset Churn Ready Topics
	err = k.ResetChurnReadyTopics(ctx)
	if err != nil {
		fmt.Println("Error resetting churn ready topics: ", err)
		return nil, cosmosMath.Uint{}, err
	}
	// Determine how many funds to draw from demand and Remove depleted/insufficiently funded requests
	totalFundsToDrawFromDemand := cosmosMath.NewUint(0)
	var topicsToSetChurn []*types.Topic
	for _, topicId := range topTopicsByReturn {
		// Add to the fee revenue collected for this topic for this reward epoch
		k.AddTopicFeeRevenue(ctx, topicId, topicBestPrices[topicId].Return)

		// Draw demand from the valid requests
		bestPrice := topicBestPrices[topicId].Price
		numRequestsServed := 0
		for _, req := range requestsToDrawDemandFrom[topicId] {
			reqDemand, err := k.GetRequestDemand(ctx, req.Id)
			if err != nil {
				fmt.Println("Error getting request demand: ", err)
				return nil, cosmosMath.Uint{}, err
			}
			// all the previous conditionals were already applied to the requests in the previous loop
			// => should never be negative
			newReqDemand := reqDemand.Sub(bestPrice)
			k.SetRequestDemand(ctx, req.Id, newReqDemand)
			// if the request is a one-shot request, remove it from the mempool

			if req.Cadence == 0 {
				k.RemoveFromMempool(ctx, req.Id)
			} else { // if it is a subscription check that the subscription has enough funds left to be worth serving
				minRequestUnmetDemand, err := k.GetParamsMinRequestUnmetDemand(ctx)
				if err != nil {
					fmt.Println("Error getting min request unmet demand: ", err)
					return nil, cosmosMath.Uint{}, err
				}
				if newReqDemand.LT(minRequestUnmetDemand) {
					// Should convey to users to not surprise them. This helps prevent spamming the mempool with requests that are not worth serving
					// The effectively burned dust is 1-time "cost" the consumer incurs when they create "subscriptions" they don't ever refill nor fill enough
					// This encourages consumers to maximize how much they fund any single request, discouraging a pattern of many less-funded requests
					k.RemoveFromMempool(ctx, req.Id)
				}
			}
			numRequestsServed++
		}
		totalFundsToDrawFromDemand = totalFundsToDrawFromDemand.Add(bestPrice.Mul(cosmosMath.NewUint(uint64(numRequestsServed))))
		topic, err := k.GetTopic(ctx, topicId)
		if err != nil {
			fmt.Println("Error getting topic: ", err)
			return nil, cosmosMath.Uint{}, err
		}
		topicsToSetChurn = append(topicsToSetChurn, &topic)
	}

	// Set the topics as churn ready
	err = k.SetChurnReadyTopics(ctx, types.TopicList{Topics: topicsToSetChurn})
	if err != nil {
		fmt.Println("Error setting churn ready topic: ", err)
		return nil, cosmosMath.Uint{}, err
	}

	return topicsToSetChurn, totalFundsToDrawFromDemand, nil
}
