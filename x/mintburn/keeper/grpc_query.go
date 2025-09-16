package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/maany-xyz/maany-provider/x/mintburn/types"
)

// Ensure queryServer implements the generated QueryServer.
var _ types.QueryServer = queryServer{}

// NewQueryServer returns a types.QueryServer backed by the given keeper.
func NewQueryServer(k Keeper) types.QueryServer {
	return queryServer{k}
}

// Escrow returns a single escrow by (consumer_chain_id, denom).
func (q queryServer) Escrow(ctx context.Context, req *types.QueryEscrowRequest) (*types.QueryEscrowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ConsumerChainId == "" || req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "consumer_chain_id and denom are required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	e, ok := q.GetEscrow(sdkCtx, req.ConsumerChainId, req.Denom)
	if !ok {
		// NotFound is also acceptable, but empty response is fine if that's your convention:
		return &types.QueryEscrowResponse{}, nil
	}
	return &types.QueryEscrowResponse{Escrow: &e}, nil
}

// Escrows lists all escrows, optionally filtered by status.
func (q queryServer) Escrows(ctx context.Context, req *types.QueryEscrowsRequest) (*types.QueryEscrowsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	list := make([]*types.Escrow, 0, 64)
	q.IterateEscrows(sdkCtx, func(e types.Escrow) (stop bool) {
		if req != nil && req.StatusFilter != "" {
			switch req.StatusFilter {
			case "PENDING":
				if e.Status != types.EscrowStatus_ESCROW_STATUS_PENDING {
					return false
				}
			case "CLAIMED":
				if e.Status != types.EscrowStatus_ESCROW_STATUS_CLAIMED {
					return false
				}
			case "CANCELED":
				if e.Status != types.EscrowStatus_ESCROW_STATUS_CANCELED {
					return false
				}
			}
		}
		ec := e // avoid taking address of loop var
		list = append(list, &ec)
		return false
	})

	return &types.QueryEscrowsResponse{Escrows: list}, nil
}

// AuthorizedICA returns the registered ICA address for a consumer chain, if set
func (q queryServer) AuthorizedICA(ctx context.Context, req *types.QueryAuthorizedICARequest) (*types.QueryAuthorizedICAResponse, error) {
    if req == nil || req.ConsumerChainId == "" {
        return nil, status.Error(codes.InvalidArgument, "consumer_chain_id is required")
    }
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    addr, found := q.GetAuthorizedICA(sdkCtx, req.ConsumerChainId)
    if !found {
        return &types.QueryAuthorizedICAResponse{IcaAddress: "", Found: false}, nil
    }
    return &types.QueryAuthorizedICAResponse{IcaAddress: addr, Found: true}, nil
}
