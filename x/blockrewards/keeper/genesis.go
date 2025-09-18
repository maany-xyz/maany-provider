package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/api/tendermint/abci"
	"github.com/maany-xyz/maany-provider/x/blockrewards/types"
)

func (k Keeper) InitGenesis(ctx context.Context, genState *types.GenesisState) []abci.ValidatorUpdate {
    // Initialize module parameters
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    // Map protobuf genesis params to runtime ModuleParams; keep default halving interval
    mp := types.DefaultModuleParams()
    if genState != nil {
        mp.RewardAmount = genState.Params.BlockRewardAmount
    }
    k.SetModuleParams(sdkCtx, mp)
    // Return validator updates if this module affects staking/validators
    return []abci.ValidatorUpdate{}
}

func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
    // Retrieve the module's parameters
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    mp := k.GetModuleParams(sdkCtx)
    return &types.GenesisState{Params: types.Params{BlockRewardAmount: mp.RewardAmount}}
}
