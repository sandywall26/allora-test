package types

import "cosmossdk.io/errors"

var (
	ErrTopicReputerStakeDoesNotExist = errors.Register(
		ModuleName,
		1,
		"topic reputer stake does not exist",
	)
	ErrIntegerUnderflowTopicReputerStake = errors.Register(
		ModuleName,
		2,
		"integer underflow for topic reputer stake",
	)
	ErrIntegerUnderflowTopicStake = errors.Register(
		ModuleName,
		3,
		"integer underflow for topic stake",
	)
	ErrIntegerUnderflowTotalStake = errors.Register(
		ModuleName,
		4,
		"integer underflow for total stake",
	)
	ErrIterationLengthDoesNotMatch = errors.Register(
		ModuleName,
		5,
		"iteration length does not match",
	)
	ErrInvalidTopicId                  = errors.Register(ModuleName, 6, "invalid topic ID")
	ErrReputerAlreadyRegisteredInTopic = errors.Register(
		ModuleName,
		7,
		"reputer already registered in topic",
	)
	ErrAddressAlreadyRegisteredInATopic = errors.Register(
		ModuleName,
		8,
		"address already registered in a topic",
	)
	ErrAddressIsNotRegisteredInAnyTopic = errors.Register(
		ModuleName,
		9,
		"address is not registered in any topic",
	)
	ErrAddressIsNotRegisteredInThisTopic = errors.Register(
		ModuleName,
		10,
		"address is not registered in this topic",
	)
	ErrInsufficientStakeToRegister = errors.Register(
		ModuleName,
		11,
		"insufficient stake to register",
	)
	ErrLibP2PKeyRequired = errors.Register(
		ModuleName,
		12,
		"libp2p key required",
	)
	ErrAddressNotRegistered = errors.Register(
		ModuleName,
		13,
		"address not registered",
	)
	ErrInsufficientStakeToRemove = errors.Register(
		ModuleName,
		14,
		"insufficient stake to remove",
	)
	ErrBlockHeightNegative = errors.Register(
		ModuleName,
		15,
		"block height negative",
	)
	ErrBlockHeightLessThanPrevious = errors.Register(
		ModuleName,
		16,
		"block height less than previous",
	)
	ErrConfirmRemoveStakeNoRemovalStarted = errors.Register(
		ModuleName,
		17,
		"confirm remove stake no removal started",
	)
	ErrConfirmRemoveStakeTooEarly = errors.Register(
		ModuleName,
		18,
		"confirm remove stake too early",
	)
	ErrConfirmRemoveStakeTooLate = errors.Register(
		ModuleName,
		19,
		"confirm remove stake too late",
	)
	ErrTopicIdListValueDecodeInvalidLength = errors.Register(
		ModuleName,
		20,
		"topic ID list value decode invalid length",
	)
	ErrTopicIdListValueDecodeJsonInvalidLength = errors.Register(
		ModuleName,
		21,
		"topic ID list value decode JSON invalid length",
	)
	ErrTopicIdListValueDecodeJsonInvalidFormat = errors.Register(
		ModuleName,
		22,
		"topic ID list value decode JSON invalid format",
	)
	ErrTopicDoesNotExist = errors.Register(
		ModuleName,
		23,
		"topic does not exist",
	)
	ErrOwnerCannotBeEmpty = errors.Register(
		ModuleName,
		24,
		"owner cannot be empty",
	)
	ErrInsufficientStakeAfterRemoval = errors.Register(
		ModuleName,
		25,
		"insufficient stake after removal",
	)
	ErrFundAmountTooLow = errors.Register(
		ModuleName,
		26,
		"topic fund amount too low",
	)
	ErrIntegerUnderflowUnmetDemand = errors.Register(
		ModuleName,
		27,
		"integer underflow for unmet demand",
	)
	ErrNotWhitelistAdmin = errors.Register(
		ModuleName,
		28,
		"not whitelist admin",
	)
	ErrNotInReputerWhitelist = errors.Register(
		ModuleName,
		30,
		"not in reputer whitelist",
	)
	ErrTopicNotEnoughDemand = errors.Register(
		ModuleName,
		31,
		"topic not enough demand",
	)
	ErrIntegerUnderflowStakeFromDelegator = errors.Register(
		ModuleName,
		32,
		"integer underflow for stake from delegator",
	)
	ErrIntegerUnderflowDelegateStakePlacement = errors.Register(
		ModuleName,
		33,
		"integer underflow for delegate stake placement",
	)
	ErrIntegerUnderflowDelegateStakeUponReputer = errors.Register(
		ModuleName,
		34,
		"integer underflow for delegate stake upon reputer",
	)
	ErrAdjustedStakeInvalidSliceLength = errors.Register(
		ModuleName,
		35,
		"adjusted stake: invalid slice length",
	)
	ErrFractionDivideByZero = errors.Register(
		ModuleName,
		36,
		"fraction: divide by zero",
	)
	ErrNumberRatioDivideByZero = errors.Register(
		ModuleName,
		37,
		"number ratio: divide by zero",
	)
	ErrNumberRatioInvalidSliceLength = errors.Register(
		ModuleName,
		38,
		"number ratio: invalid slice length",
	)
	ErrInvalidSliceLength = errors.Register(
		ModuleName,
		39,
		"invalid slice length",
	)
	ErrTopicCadenceBelowMinimum = errors.Register(
		ModuleName,
		40,
		"topic cadence must be at least 60 seconds (1 minute)",
	)
	ErrPhiCannotBeZero = errors.Register(
		ModuleName,
		41,
		"phi: cannot be zero",
	)
	ErrSumWeightsLessThanEta = errors.Register(
		ModuleName,
		42,
		"sum weights less than eta",
	)
	ErrSliceLengthMismatch = errors.Register(
		ModuleName,
		43,
		"slice length mismatch",
	)
	ErrNonceAlreadyFulfilled = errors.Register(
		ModuleName,
		44,
		"nonce already fulfilled",
	)
	ErrNonceStillUnfulfilled = errors.Register(
		ModuleName,
		45,
		"nonce still unfulfilled",
	)
	ErrTopicCreatorNotEnoughDenom = errors.Register(
		ModuleName,
		46,
		"topic creator does not have enough denom",
	)
	ErrSignatureVerificationFailed = errors.Register(
		ModuleName,
		47,
		"signature verification was failed",
	)
	ErrTopicRegistrantNotEnoughDenom = errors.Register(
		ModuleName,
		48,
		"topic registrant does not have enough denom",
	)
	ErrInsufficientDelegateStakeToRemove = errors.Register(
		ModuleName,
		49,
		"insufficient delegate stake to remove",
	)
	ErrTopicMempoolAtCapacity = errors.Register(
		ModuleName,
		50,
		"topic mempool at capacity",
	)
	ErrReceivedZeroAmount = errors.Register(
		ModuleName,
		51,
		"received zero amount",
	)
	ErrInvalidValue  = errors.Register(ModuleName, 52, "invalid value")
	ErrQueryTooLarge = errors.Register(
		ModuleName,
		53,
		"query is too large",
	)
	ErrFailedToSerializePayload = errors.Register(
		ModuleName,
		54,
		"failed to serialize payload",
	)
	ErrNoValidBundles = errors.Register(
		ModuleName,
		55,
		"no valid bundles found",
	)
	ErrInvalidReward    = errors.Register(ModuleName, 56, "invalid reward")
	ErrCantSelfDelegate = errors.Register(
		ModuleName,
		57,
		"cant self delegate",
	)
	ErrValidationVersionEmpty = errors.Register(
		ModuleName,
		58,
		"version cannot be empty",
	)
	ErrValidationVersionTooLong = errors.Register(
		ModuleName,
		59,
		"version length cannot exceed 32 characters",
	)
	ErrValidationMustBeNonNegative = errors.Register(
		ModuleName,
		60,
		"value must be non-negative",
	)
	ErrValidationMustBeBetweenZeroAndOne = errors.Register(
		ModuleName,
		61,
		"value must be between 0 and 1",
	)
	ErrValidationMustBeGreaterthanZero = errors.Register(
		ModuleName,
		62,
		"value must be greater than 0",
	)
)
