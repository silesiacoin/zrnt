package meta

import (
	"github.com/protolambda/zrnt/eth2/beacon/validator"
	. "github.com/protolambda/zrnt/eth2/core"
)

type ExitMeta interface {
	InitiateValidatorExit(currentEpoch Epoch, index ValidatorIndex)
}

type BalanceMeta interface {
	GetBalance(index ValidatorIndex) Gwei
	IncreaseBalance(index ValidatorIndex, v Gwei)
	DecreaseBalance(index ValidatorIndex, v Gwei)
}

type RegistrySizeMeta interface {
	IsValidIndex(index ValidatorIndex) bool
	ValidatorCount() uint64
}

type PubkeyMeta interface {
	Pubkey(index ValidatorIndex) BLSPubkey
	ValidatorIndex(pubkey BLSPubkey) (index ValidatorIndex, exists bool)
}

type EffectiveBalanceMeta interface {
	EffectiveBalance(index ValidatorIndex) Gwei
	SumEffectiveBalanceOf(indices []ValidatorIndex) (sum Gwei)
}

type FinalityMeta interface {
	Finalized() Checkpoint
	CurrentJustified() Checkpoint
	PreviousJustified() Checkpoint
}

type AttesterStatusMeta interface {
	GetAttesterStatus(index ValidatorIndex) AttesterStatus
}

type SlashedMeta interface {
	IsSlashed(i ValidatorIndex) bool
	FilterUnslashed(indices []ValidatorIndex) []ValidatorIndex
}

type CompactValidatorMeta interface {
	PubkeyMeta
	EffectiveBalanceMeta
	SlashedMeta
}

type StakingMeta interface {
	// Staked = Active effective balance
	GetTotalStakedBalance(epoch Epoch) Gwei
}

type FFGMeta interface {
	StakingMeta
	GetTargetTotalStakedBalance(epoch Epoch) Gwei
}

type SlashingMeta interface {
	GetIndicesToSlash(withdrawal Epoch) (out []ValidatorIndex)
}

type ValidatorMeta interface {
	Validator(index ValidatorIndex) *validator.Validator
}

type VersioningMeta interface {
	CurrentSlot() Slot
	CurrentEpoch() Epoch
	PreviousEpoch() Epoch
	GetDomain(dom BLSDomainType, messageEpoch Epoch) BLSDomain
}

type Eth1Meta interface {
	DepIndex() DepositIndex
	DepCount() DepositIndex
	DepRoot() Root
}

type OnboardMeta interface {
	AddNewValidator(pubkey BLSPubkey, withdrawalCreds Root, balance Gwei)
}

type DepositMeta interface {
	BalanceMeta
	OnboardMeta
	IncrementDepositIndex()
}

type HeaderMeta interface {
	// Signing root of latest_block_header
	GetLatestBlockRoot() Root
}

type UpdateHeaderMeta interface {
	StoreHeaderData(slot Slot, parent Root, body Root)
	UpdateLatestBlockRoot(stateRoot Root) Root
}

type HistoryMeta interface {
	GetBlockRootAtSlot(slot Slot) Root
	GetBlockRoot(epoch Epoch) Root
}

type HistoryUpdateMeta interface {
	SetRecentRoots(slot Slot, blockRoot Root, stateRoot Root)
	UpdateHistoricalRoots()
}

type SeedMeta interface {
	// Retrieve the seed for beacon proposer indices.
	GetSeed(epoch Epoch) Root
}

type ProposingMeta interface {
	GetBeaconProposerIndex() ValidatorIndex
}

type ActivationExitMeta interface {
	GetChurnLimit(epoch Epoch) uint64
	ExitQueueEnd(epoch Epoch) Epoch
}

type ActiveValidatorCountMeta interface {
	GetActiveValidatorCount(epoch Epoch) uint64
}

type ActiveIndicesMeta interface {
	GetActiveValidatorIndices(epoch Epoch) []ValidatorIndex
}

type ActiveIndexRootMeta interface {
	GetActiveIndexRoot(epoch Epoch) Root
	// TODO
	// indices := state.Validators.GetActiveValidatorIndices(epoch)
	// ssz.HashTreeRoot(indices, RegistryIndicesSSZ)
}

type CommitteeCountMeta interface {
	// Amount of committees per epoch. Minimum is SLOTS_PER_EPOCH
	GetCommitteeCount(epoch Epoch) uint64
}

type CrosslinkTimingMeta interface {
	GetStartShard(epoch Epoch) Shard
}

type CrosslinkCommitteeMeta interface {
	GetCrosslinkCommittee(epoch Epoch, shard Shard) []ValidatorIndex
}

type CrosslinkMeta interface {
	GetCurrentCrosslinkRoots() *[SHARD_COUNT]Root
	GetPreviousCrosslinkRoots() *[SHARD_COUNT]Root
	GetCurrentCrosslink(shard Shard) *Crosslink
	GetPreviousCrosslink(shard Shard) *Crosslink
}

type WinningCrosslinkMeta interface {
	GetWinningCrosslinkAndAttesters(epoch Epoch, shard Shard) (*Crosslink, ValidatorSet)
}

type RandomnessMeta interface {
	GetRandomMix(epoch Epoch) Root
	// epoch + EPOCHS_PER_HISTORICAL_VECTOR - MIN_SEED_LOOKAHEAD) // Avoid underflow
}
