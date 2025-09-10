package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/api/tendermint/abci"
	"github.com/maany-xyz/maany-provider/x/blockrewards/types"
)

func (k Keeper) InitGenesis(ctx context.Context, genState *types.GenesisState) []abci.ValidatorUpdate {
    // Initialize module parameters
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    if err := k.SetParams(sdkCtx, genState.Params); err != nil {
        panic(fmt.Sprintf("failed to set params in InitGenesis: %v", err))
    }
    // Return validator updates if this module affects staking/validators
    return []abci.ValidatorUpdate{}
}

func (k Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
    // Retrieve the module's parameters
    sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, err := k.GetParams(sdkCtx)
    if err != nil {
        return nil
    }
    return &types.GenesisState{
        Params: params,
    }
}