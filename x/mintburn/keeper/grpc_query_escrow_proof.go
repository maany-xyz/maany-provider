// x/mintburn/keeper/grpc_query_escrow_proof.go
package keeper

import (
	"context"
	"encoding/hex"

	// SDK v0.50 import path
	sdk "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/maany-xyz/maany-provider/x/mintburn/types"
)

// queryServer implements types.QueryServer.
type queryServer struct {
	Keeper
}

var _ types.QueryServer = queryServer{}

// EscrowProof returns the membership inputs needed by a consumer to verify
// an escrow record at a given height: the exact committed VALUE bytes,
// the exact KEY path segments, and the block height used.
//
// IMPORTANT: On Cosmos SDK v0.50+, building ICS-23 proofs inside a module
// is not supported (MultiStore.GetCommitmentProof was removed).
// Fetch the proof via ABCI RPC (/abci_query?prove=true) off-chain using the
// key returned here, then combine it with this response for your consumer genesis.
func (q queryServer) EscrowProof(ctx context.Context, req *types.QueryEscrowProofRequest) (*types.QueryEscrowProofResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ConsumerChainId == "" || req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "consumer_chain_id and denom are required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Only the *current* block height is available inside ABCI context.
	// For genesis bootstrap that is fine: do escrow, then query next block.
	height := uint64(sdkCtx.BlockHeight())
	if req.Height != 0 && req.Height != height {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"only current block height supported from in-module query; requested=%d current=%d",
			req.Height, height,
		)
	}

	// 1) Lookup Escrow by (consumer_chain_id, denom)
	esc, found := q.GetEscrow(sdkCtx, req.ConsumerChainId, req.Denom)
	if !found {
		return nil, status.Errorf(codes.NotFound, "escrow not found for consumer_chain_id=%s denom=%s", req.ConsumerChainId, req.Denom)
	}
	if esc.EscrowId == "" {
		return nil, status.Error(codes.Internal, "escrow_id missing on escrow record")
	}

	// 2) Marshal the exact committed value bytes (must match stored bytes)
	valueBz := q.cdc.MustMarshal(&esc)

	// 3) Build the key path segments expected by consumer verification.
	// Canonical substore name:
	substore := types.StoreKey // e.g. "mintburn" or "x-mintburn" (use your actual StoreKey)
	// Primary key under that substore:
	key := types.EscrowKeyByID(esc.EscrowId)
	keyHex := hex.EncodeToString(key)

	// The consumer will verify with NewMerklePath(substore, keyHex)
	keyPath := []string{substore, keyHex}

	// 4) Return response (without Merkle proof; fetch via RPC off-chain)
	resp := &types.QueryEscrowProofResponse{
		Height:    height,
		Value:     valueBz,
		KeyPath:   keyPath,
		EscrowId:  esc.EscrowId,
		AmountDenom: esc.Amount.Denom,
		AmountValue: esc.Amount.Amount.String(),
		// MerkleProof: nil  // intentionally omitted; fetch over RPC
	}
	return resp, nil
}

