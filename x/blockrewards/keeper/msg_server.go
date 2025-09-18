package keeper

import (
    "context"

    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    "github.com/maany-xyz/maany-provider/x/blockrewards/types"
)

// Ensure Keeper satisfies the generated MsgServer interface.
var _ types.MsgServer = Keeper{}

func (k Keeper) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Authority check
    if msg.Authority != k.authority {
        return nil, sdkerrors.ErrUnauthorized
    }

    // Build new params from message
    next := types.ModuleParams{
        RewardAmount:    msg.RewardAmount,
        HalvingInterval: msg.HalvingInterval,
    }
    if err := next.Validate(); err != nil {
        return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
    }

    k.SetModuleParams(ctx, next)
    return &types.MsgUpdateParamsResponse{}, nil
}

