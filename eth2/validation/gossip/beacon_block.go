package gossip

import (
	"context"
	"errors"
	"fmt"
	"github.com/protolambda/zrnt/eth2/beacon"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/ztyp/tree"
)

// Checks if the (slot, proposer) pair was seen, does not do any tracking.
type BlockSeenFn func(slot beacon.Slot, proposer beacon.ValidatorIndex) bool

// When the block is fully validated (except proposer index check, but incl. signature check),
// the combination can be marked as seen to avoid future duplicate blocks from being propagated.
// Must returns true if the block was previously already seen,
// to avoid race-conditions (two block validations may run in parallel).
type MarkBlockSeenFn func(slot beacon.Slot, proposer beacon.ValidatorIndex) bool

func (gv *GossipValidator) ValidateBeaconBlock(ctx context.Context, signedBlock *beacon.SignedBeaconBlock,
	hasBlock BlockSeenFn, markBlock MarkBlockSeenFn) GossipValidatorResult {

	block := &signedBlock.Message
	// [IGNORE] The block is not from a future slot (with a MAXIMUM_GOSSIP_CLOCK_DISPARITY allowance) --
	// i.e. validate that signed_beacon_block.message.slot <= current_slot
	if maxSlot := gv.SlotAfter(MAXIMUM_GOSSIP_CLOCK_DISPARITY); maxSlot < block.Slot {
		return GossipValidatorResult{IGNORE, fmt.Errorf("block slot %d is later than max slot %d", block.Slot, maxSlot)}
	}

	// [IGNORE] The block is the first block with valid signature received for the proposer for the slot, signed_beacon_block.message.slot.
	if hasBlock(block.Slot, block.ProposerIndex) {
		return GossipValidatorResult{IGNORE, fmt.Errorf("already seen a block for slot %d proposer %d", block.Slot, block.ProposerIndex)}
	}

	// [IGNORE] The block's parent (defined by block.parent_root) has been seen
	// (via both gossip and non-gossip sources)
	parentRef, err := gv.Chain.ByBlockRoot(block.ParentRoot)
	if err != nil {
		return GossipValidatorResult{IGNORE, fmt.Errorf("block has unavailable parent block %s: %w", block.ParentRoot, err)}
	}
	// Sanity check, implied condition
	if parentRef.Slot() >= block.Slot {
		// It's OK to propagate, others do so and attack scope is limited, but it will not be processed later on. So just ignore it.
		return GossipValidatorResult{IGNORE, fmt.Errorf("block slot %d not after parent %d (%s)", block.Slot, parentRef.Slot(), block.ParentRoot)}
	}

	// [IGNORE] The block is from a slot greater than the latest finalized slot --
	// i.e. validate that signed_beacon_block.message.slot > compute_start_slot_at_epoch(state.finalized_checkpoint.epoch)
	fin := gv.Chain.Finalized()
	if finSlot, _ := gv.Spec.EpochStartSlot(fin.Epoch); block.Slot <= finSlot {
		return GossipValidatorResult{IGNORE, fmt.Errorf("block slot %d is not after finalized slot %d", block.Slot, finSlot)}
	}
	// [REJECT] The current finalized_checkpoint is an ancestor of block -- i.e. get_ancestor(store, block.parent_root, compute_start_slot_at_epoch(store.finalized_checkpoint.epoch)) == store.finalized_checkpoint.root
	if unknown, isAncestor := gv.Chain.IsAncestor(block.ParentRoot, fin.Root); unknown {
		return GossipValidatorResult{IGNORE, fmt.Errorf("failed to determine if parent block %s is in subtree of finalized block %s", block.ParentRoot, fin.Root)}
	} else if !isAncestor && block.ParentRoot != fin.Root { // If it builds on the finalized block itself, that is still ok.
		return GossipValidatorResult{REJECT, fmt.Errorf("parent block %s is not in subtree of finalized root %s", block.ParentRoot, fin.Root)}
	}

	// [REJECT] The block's parent (defined by block.parent_root) passes validation.
	// *implicit*: parent was already processed and put into forkchoice view, so it passes validation.

	parentEpc, err := parentRef.EpochsContext(ctx)
	if err != nil {
		return GossipValidatorResult{IGNORE, fmt.Errorf("cannot find context for parent block %s", block.ParentRoot)}
	}
	// [REJECT] The proposer signature, signed_beacon_block.signature, is valid with respect to the proposer_index pubkey.
	pub, ok := parentEpc.PubkeyCache.Pubkey(block.ProposerIndex)
	if !ok {
		return GossipValidatorResult{IGNORE, fmt.Errorf("cannot find pubkey for proposer index %d", block.ProposerIndex)}
	}
	domain, err := gv.GetDomain(gv.Spec.DOMAIN_BEACON_PROPOSER, gv.Spec.SlotToEpoch(block.Slot))
	if err != nil {
		return GossipValidatorResult{IGNORE, fmt.Errorf("cannot get signature domain for block at slot %d", block.Slot)}
	}
	if bls.Verify(pub, beacon.ComputeSigningRoot(block.HashTreeRoot(gv.Spec, tree.GetHashFn()), domain), signedBlock.Signature) {
		return GossipValidatorResult{REJECT, errors.New("invalid block signature")}
	}

	if !markBlock(block.Slot, block.ProposerIndex) {
		// since we last checked, a parallel validation may have accepted a block for this (slot, proposer) pair. So we do this validation again.
		return GossipValidatorResult{IGNORE, fmt.Errorf("already seen a block for slot %d proposer %d", block.Slot, block.ProposerIndex)}
	}

	// [REJECT] The block is proposed by the expected proposer_index for the block's slot in the context of
	// the current shuffling (defined by parent_root/slot).

	// TODO: get target epoch EPC from chain, with potential necessary slot transitioning. Transitioning is missing now.
	targetEpoch := gv.Spec.SlotToEpoch(block.Slot)
	parentEpoch := gv.Spec.SlotToEpoch(parentRef.Slot())
	var proposer beacon.ValidatorIndex
	if targetEpoch == parentEpoch {
		proposer, err = parentEpc.GetBeaconProposer(block.Slot)
		if err != nil {
			return GossipValidatorResult{IGNORE, fmt.Errorf("could not get proposer index for slot %d, from same epoch as parent block", block.Slot)}
		}
	} else {
		// TODO There's transitioning to do.
		proposer = 1234
	}

	if proposer != block.ProposerIndex {
		return GossipValidatorResult{REJECT, fmt.Errorf("expected proposer %d, but block was proposed by %d", proposer, block.ProposerIndex)}
	}

	return GossipValidatorResult{ACCEPT, nil}
}