package rewards

import (
	"math"

	"github.com/allora-network/allora-chain/x/emissions/keeper"
	"github.com/allora-network/allora-chain/x/emissions/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

/*
 These functions will be used immediately after the network loss for the relevant time step has been generated.
 Using the network loss and the sets of losses reported by each repeater, the scores are calculated. In the case
 of workers (who perform the forecast task and network task), the last 10 previous scores will also be taken into
 consideration to generate the score at the most recent time step.
*/

// GenerateReputerScores calculates and persists scores for reputers based on their reported losses.
func GenerateReputerScores(ctx sdk.Context, keeper keeper.Keeper, topicId uint64, block int64, reportedLosses types.ReputerValueBundles) ([]types.Score, error) {
	// Fetch all workers for the topic to determine expected report coverage
	allWorkers, err := keeper.GetTopicWorkers(ctx, topicId)
	if err != nil {
		return nil, err
	}
	allWorkerAddresses := make(map[string]struct{})
	for _, worker := range allWorkers {
		allWorkerAddresses[worker.String()] = struct{}{}
	}

	// Fetch reputers data
	var reputerAddresses []sdk.AccAddress
	var reputerStakes []float64
	var reputerListeningCoefficients []float64
	var losses [][]float64
	for _, reportedLoss := range reportedLosses.ReputerValueBundles {
		reputerAddr, err := sdk.AccAddressFromBech32(reportedLoss.Reputer)
		if err != nil {
			return []types.Score{}, err
		}
		reputerAddresses = append(reputerAddresses, reputerAddr)

		// Get reputer topic stake
		reputerStake, err := keeper.GetStakeOnTopicFromReputer(ctx, topicId, reputerAddr)
		if err != nil {
			return []types.Score{}, err
		}
		reputerStakes = append(reputerStakes, float64(reputerStake.BigInt().Int64()))

		// Get reputer listening coefficient
		res, err := keeper.GetListeningCoefficient(ctx, topicId, reputerAddr)
		if err != nil {
			return []types.Score{}, err
		}
		reputerListeningCoefficients = append(reputerListeningCoefficients, res.Coefficient)

		// Check if all workers are reported in OneOutInfererValues, OneOutForecasterValues, OneInForecasterValues
		workerOneOutInfererValues := map[string]bool{}
		workerOneOutForecasterValues := map[string]bool{}
		workerOneInForecasterValues := map[string]bool{}

		// Mark workers reported in OneOutInfererValues
		for _, value := range reportedLoss.ValueBundle.OneOutInfererValues {
			if _, exists := allWorkerAddresses[value.Worker]; exists {
				workerOneOutInfererValues[value.Worker] = true
			}
		}
		// Mark workers reported in OneOutForecasterValues
		for _, value := range reportedLoss.ValueBundle.OneOutForecasterValues {
			if _, exists := allWorkerAddresses[value.Worker]; exists {
				workerOneOutForecasterValues[value.Worker] = true
			}
		}
		// Mark workers reported in OneInForecasterValues
		for _, value := range reportedLoss.ValueBundle.OneInForecasterValues {
			if _, exists := allWorkerAddresses[value.Worker]; exists {
				workerOneInForecasterValues[value.Worker] = true
			}
		}

		// Check the missing workers and add NaN values for the not reported
		// This will help identify the losses that were not reported by the reputer
		for worker := range allWorkerAddresses {
			if _, reported := workerOneOutInfererValues[worker]; !reported {
				reportedLoss.ValueBundle.OneOutInfererValues = append(reportedLoss.ValueBundle.OneOutInfererValues, &types.WithheldWorkerAttributedValue{Worker: worker, Value: math.NaN()})
			}
			if _, reported := workerOneOutForecasterValues[worker]; !reported {
				reportedLoss.ValueBundle.OneOutForecasterValues = append(reportedLoss.ValueBundle.OneOutForecasterValues, &types.WithheldWorkerAttributedValue{Worker: worker, Value: math.NaN()})
			}
			if _, reported := workerOneInForecasterValues[worker]; !reported {
				reportedLoss.ValueBundle.OneInForecasterValues = append(reportedLoss.ValueBundle.OneInForecasterValues, &types.WorkerAttributedValue{Worker: worker, Value: math.NaN()})
			}
		}

		// Get all reported losses from bundle
		reputerLosses := ExtractValues(reportedLoss.ValueBundle)
		losses = append(losses, reputerLosses)
	}

	// Get reputer output
	scores, newCoefficients, err := GetAllReputersOutput(losses, reputerStakes, reputerListeningCoefficients, len(reputerStakes))
	if err != nil {
		return []types.Score{}, err
	}

	// Insert new coeffients and scores
	var newScores []types.Score
	for i, reputerAddr := range reputerAddresses {
		err := keeper.SetListeningCoefficient(ctx, topicId, reputerAddr, types.ListeningCoefficient{Coefficient: newCoefficients[i]})
		if err != nil {
			return []types.Score{}, err
		}

		newScore := types.Score{
			TopicId:     topicId,
			BlockNumber: block,
			Address:     reputerAddr.String(),
			Score:       scores[i],
		}
		err = keeper.InsertReputerScore(ctx, topicId, block, newScore)
		if err != nil {
			return []types.Score{}, err
		}
		newScores = append(newScores, newScore)
	}

	return newScores, nil
}

// GenerateInferenceScores calculates and persists scores for workers based on their inference task performance.
func GenerateInferenceScores(ctx sdk.Context, keeper keeper.Keeper, topicId uint64, block int64, networkLosses types.ValueBundle) ([]types.Score, error) {
	var newScores []types.Score
	for _, oneOutLoss := range networkLosses.OneOutInfererValues {
		workerAddr, err := sdk.AccAddressFromBech32(oneOutLoss.Worker)
		if err != nil {
			return []types.Score{}, err
		}

		// Calculate new score
		workerNewScore := GetWorkerScore(networkLosses.CombinedValue, oneOutLoss.Value)

		newScore := types.Score{
			TopicId:     topicId,
			BlockNumber: block,
			Address:     workerAddr.String(),
			Score:       workerNewScore,
		}
		err = keeper.InsertWorkerInferenceScore(ctx, topicId, block, newScore)
		if err != nil {
			return []types.Score{}, err
		}
		newScores = append(newScores, newScore)
	}
	return newScores, nil
}

// GenerateForecastScores calculates and persists scores for workers based on their forecast task performance.
func GenerateForecastScores(ctx sdk.Context, keeper keeper.Keeper, topicId uint64, block int64, networkLosses types.ValueBundle) ([]types.Score, error) {
	// Get worker scores for one out loss
	var workersScoresOneOut []float64
	for _, oneOutLoss := range networkLosses.OneOutForecasterValues {
		workerScore := GetWorkerScore(networkLosses.CombinedValue, oneOutLoss.Value)
		workersScoresOneOut = append(workersScoresOneOut, workerScore)
	}

	numForecasters := len(workersScoresOneOut)
	fUniqueAgg := GetfUniqueAgg(float64(numForecasters))
	var newScores []types.Score
	for i, oneInNaiveLoss := range networkLosses.OneInForecasterValues {
		workerAddr, err := sdk.AccAddressFromBech32(oneInNaiveLoss.Worker)
		if err != nil {
			return []types.Score{}, err
		}

		// Get worker score for one in loss
		workerScoreOneIn := GetWorkerScore(oneInNaiveLoss.Value, networkLosses.NaiveValue)

		// Calculate forecast score
		workerFinalScore := GetFinalWorkerScoreForecastTask(workerScoreOneIn, workersScoresOneOut[i], fUniqueAgg)

		newScore := types.Score{
			TopicId:     topicId,
			BlockNumber: block,
			Address:     workerAddr.String(),
			Score:       workerFinalScore,
		}
		err = keeper.InsertWorkerForecastScore(ctx, topicId, block, newScore)
		if err != nil {
			return []types.Score{}, err
		}
		newScores = append(newScores, newScore)
	}

	return newScores, nil
}
