package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/maany-xyz/maany-provider/x/blockrewards/types"
)

// NewQueryServer returns a QueryServer backed by keeper
func NewQueryServer(k Keeper) types.QueryServer { return k }

// Params returns the current module params
func (k Keeper) Params(goCtx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    mp := k.GetModuleParams(ctx)
    out := &types.QueryParamsResponse{
        Params: types.QueryModuleParams(mp),
    }
    return out, nil
}

// EffectiveReward computes the effective reward at the requested height
func (k Keeper) EffectiveReward(goCtx context.Context, req *types.QueryEffectiveRewardRequest) (*types.QueryEffectiveRewardResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)
    mp := k.GetModuleParams(ctx)
    denom := mp.RewardAmount.Denom
    amt := mp.RewardAmount.Amount

    height := uint64(ctx.BlockHeight())
    if req != nil && req.Height > 0 {
        height = req.Height
    }
    if mp.HalvingInterval > 0 {
        halvings := height / mp.HalvingInterval
        for i := uint64(0); i < halvings && !amt.IsZero(); i++ {
            amt = amt.QuoRaw(2)
        }
    }
    return &types.QueryEffectiveRewardResponse{Reward: sdk.NewCoin(denom, amt)}, nil
}
