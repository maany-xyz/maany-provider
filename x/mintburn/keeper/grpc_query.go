package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maany-xyz/maany-provider/x/mintburn/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Escrow(ctx context.Context, req *types.QueryEscrowRequest) (*types.QueryEscrowResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	e, ok := k.GetEscrow(sdkCtx, req.ConsumerChainId, req.Denom)
	if !ok { return &types.QueryEscrowResponse{}, nil }
	return &types.QueryEscrowResponse{Escrow: &e}, nil
}

func (k Keeper) Escrows(ctx context.Context, req *types.QueryEscrowsRequest) (*types.QueryEscrowsResponse, error) {
    sdkCtx := sdk.UnwrapSDKContext(ctx)

    list := make([]*types.Escrow, 0)
    k.IterateEscrows(sdkCtx, func(e types.Escrow) bool {
        // (optional) apply status filter
        if req != nil && req.StatusFilter != "" {
            // example: accept "PENDING", "CLAIMED", "CANCELED"
            switch req.StatusFilter {
            case "PENDING":
                if e.Status != types.EscrowStatus_ESCROW_STATUS_PENDING { return false }
            case "CLAIMED":
                if e.Status != types.EscrowStatus_ESCROW_STATUS_CLAIMED { return false }
            case "CANCELED":
                if e.Status != types.EscrowStatus_ESCROW_STATUS_CANCELED { return false }
            }
        }

        // copy to avoid taking address of iteration variable
        eCopy := e
        list = append(list, &eCopy)
        return false
    })

    return &types.QueryEscrowsResponse{Escrows: list}, nil
}

// TEMP: implement later; stub satisfies the interface
func (k Keeper) EscrowProof(ctx context.Context, req *types.QueryEscrowProofRequest) (*types.QueryEscrowProofResponse, error) {
    return nil, status.Error(codes.Unimplemented, "EscrowProof not implemented yet")
}