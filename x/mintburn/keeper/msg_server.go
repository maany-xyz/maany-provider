package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/maany-xyz/maany-provider/x/mintburn/types"
)

var _ types.MsgServer = Keeper{}

func (k Keeper) EscrowInitial(goCtx context.Context, msg *types.MsgEscrowInitial) (*types.MsgEscrowInitialResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, sdkerrors.ErrInvalidCoins
	}
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { return nil, sdkerrors.ErrInvalidAddress }
    
	modAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if err := k.bankKeeper.SendCoins(ctx, sender, modAddr, sdk.NewCoins(msg.Amount)); err != nil {
		return nil, err
	}

	esc := types.Escrow{
		ConsumerChainId: msg.ConsumerChainId,
		Amount:          msg.Amount,
		Recipient:       msg.Recipient,
		ExpiryHeight:    msg.ExpiryHeight,
		ExpiryTimeUnix:  msg.ExpiryTimeUnix,
		Status:          types.EscrowStatus_ESCROW_STATUS_PENDING,
	}
	k.SetEscrow(ctx, esc)

	return &types.MsgEscrowInitialResponse{}, nil
}

func (k Keeper) CancelEscrow(goCtx context.Context, msg *types.MsgCancelEscrow) (*types.MsgCancelEscrowResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	esc, ok := k.GetEscrow(ctx, msg.ConsumerChainId, msg.Denom)
	if !ok || esc.Status != types.EscrowStatus_ESCROW_STATUS_PENDING {
		return nil, sdkerrors.ErrNotFound
	}
	// auth/expiry checks…

	modAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	to, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil { return nil, sdkerrors.ErrInvalidAddress }
	if err := k.bankKeeper.SendCoins(ctx, modAddr, to, sdk.NewCoins(esc.Amount)); err != nil {
		return nil, err
	}

	esc.Status = types.EscrowStatus_ESCROW_STATUS_CANCELED
	k.SetEscrow(ctx, esc)
	return &types.MsgCancelEscrowResponse{}, nil
}
