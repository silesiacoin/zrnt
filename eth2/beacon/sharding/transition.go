package sharding

import (
	"context"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/zrnt/eth2/beacon/merge"
	"github.com/protolambda/zrnt/eth2/beacon/phase0"
)

func (state *BeaconStateView) ProcessEpoch(ctx context.Context, spec *common.Spec, epc *common.EpochsContext) error {
	if err := ProcessPendingShardConfirmations(ctx, spec, state); err != nil {
		return err
	}
	if err := ChargeConfirmedShardFees(ctx, spec, state, epc); err != nil {
		return err
	}
	if err := ResetPendingShardWork(ctx, spec, state, epc); err != nil {
		return err
	}

	vals, err := state.Validators()
	if err != nil {
		return err
	}
	flats, err := common.FlattenValidators(vals)
	if err != nil {
		return err
	}
	attesterData, err := phase0.ComputeEpochAttesterData(ctx, spec, epc, flats, state)
	if err != nil {
		return err
	}
	just := phase0.JustificationStakeData{
		CurrentEpoch:                  epc.CurrentEpoch.Epoch,
		TotalActiveStake:              epc.TotalActiveStake,
		PrevEpochUnslashedTargetStake: attesterData.PrevEpochUnslashedStake.TargetStake,
		CurrEpochUnslashedTargetStake: attesterData.CurrEpochUnslashedTargetStake,
	}
	if err := phase0.ProcessEpochJustification(ctx, spec, &just, state); err != nil {
		return err
	}
	if err := phase0.ProcessEpochRewardsAndPenalties(ctx, spec, epc, attesterData, state); err != nil {
		return err
	}
	if err := phase0.ProcessEpochRegistryUpdates(ctx, spec, epc, flats, state); err != nil {
		return err
	}
	if err := phase0.ProcessEpochSlashings(ctx, spec, epc, flats, state); err != nil {
		return err
	}

	if err := phase0.ProcessEth1DataReset(ctx, spec, epc, state); err != nil {
		return err
	}
	if err := phase0.ProcessEffectiveBalanceUpdates(ctx, spec, epc, flats, state); err != nil {
		return err
	}
	if err := phase0.ProcessSlashingsReset(ctx, spec, epc, state); err != nil {
		return err
	}
	if err := phase0.ProcessRandaoMixesReset(ctx, spec, epc, state); err != nil {
		return err
	}
	if err := phase0.ProcessHistoricalRootsUpdate(ctx, spec, epc, state); err != nil {
		return err
	}
	if err := phase0.ProcessParticipationRecordUpdates(ctx, spec, epc, state); err != nil {
		return err
	}
	return nil
}

func (state *BeaconStateView) ProcessBlock(ctx context.Context, spec *common.Spec, epc *common.EpochsContext, benv *common.BeaconBlockEnvelope) error {
	signedBlock, ok := benv.SignedBlock.(*SignedBeaconBlock)
	if !ok {
		return fmt.Errorf("unexpected block type %T in Merge ProcessBlock", benv.SignedBlock)
	}
	block := &signedBlock.Message
	slot, err := state.Slot()
	if err != nil {
		return err
	}
	proposerIndex, err := epc.GetBeaconProposer(slot)
	if err != nil {
		return err
	}
	if err := common.ProcessHeader(ctx, spec, state, block.Header(spec), proposerIndex); err != nil {
		return err
	}
	body := &block.Body
	if err := phase0.ProcessRandaoReveal(ctx, spec, epc, state, body.RandaoReveal); err != nil {
		return err
	}
	if err := phase0.ProcessEth1Vote(ctx, spec, epc, state, body.Eth1Data); err != nil {
		return err
	}
	// Safety checks, in case the user of the function provided too many operations
	if err := body.CheckLimits(spec); err != nil {
		return err
	}

	if err := phase0.ProcessProposerSlashings(ctx, spec, epc, state, body.ProposerSlashings); err != nil {
		return err
	}
	if err := phase0.ProcessAttesterSlashings(ctx, spec, epc, state, body.AttesterSlashings); err != nil {
		return err
	}
	// TODO: process shard proposer slashings
	// TODO: check shard headers count against active shard count
	// TODO: process shard headers
	if err := phase0.ProcessAttestations(ctx, spec, epc, state, body.Attestations); err != nil {
		return err
	}
	// TODO: wrap attestation processing, to add shard confirmation work
	if err := phase0.ProcessDeposits(ctx, spec, epc, state, body.Deposits); err != nil {
		return err
	}
	if err := phase0.ProcessVoluntaryExits(ctx, spec, epc, state, body.VoluntaryExits); err != nil {
		return err
	}
	// TODO: check if execution is enabled, like in merge?
	if err := merge.ProcessExecutionPayload(ctx, spec, state, &body.ExecutionPayload, spec.ExecutionEngine); err != nil {
		return err
	}
	//
	return nil
}

func (state *BeaconStateView) IsExecutionEnabled(spec *common.Spec, block *BeaconBlock) (bool, error) {
	return true, nil
}

func (state *BeaconStateView) IsTransitionCompleted() (bool, error) {
	return true, nil
}

func (state *BeaconStateView) IsTransitionBlock(spec *common.Spec, block *BeaconBlock) (bool, error) {
	return false, nil
}