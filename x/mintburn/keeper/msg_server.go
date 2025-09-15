package keeper

import (
	"context"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/maany-xyz/maany-provider/x/mintburn/types"
)

// Ensure Keeper satisfies the generated MsgServer.
var _ types.MsgServer = Keeper{}

// NextEscrowID returns a monotonically increasing uint64 as a string id.
// (Simple, deterministic; replace with hash if you prefer.)
func (k Keeper) NextEscrowID(ctx sdk.Context) string {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get(types.EscrowIDCounterKey) // []byte{0x00}
	var n uint64
	if len(bz) == 8 {
		n = binary.BigEndian.Uint64(bz)
	}
	n++
	out := make([]byte, 8)
	binary.BigEndian.PutUint64(out, n)
	store.Set(types.EscrowIDCounterKey, out)
	return types.Uint64ToString(n) // helper returns decimal string
}

func (k Keeper) EscrowInitial(goCtx context.Context, msg *types.MsgEscrowInitial) (*types.MsgEscrowInitialResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !msg.Amount.IsValid() || !msg.Amount.Amount.IsPositive() {
		return nil, sdkerrors.ErrInvalidCoins
	}
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}

	// move funds into module (lock)
	modAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	if err := k.bankKeeper.SendCoins(ctx, sender, modAddr, sdk.NewCoins(msg.Amount)); err != nil {
		return nil, err
	}

	// create escrow with a new escrow_id
	esc := types.Escrow{
		EscrowId:        k.NextEscrowID(ctx),
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
	if !ok {
		return nil, sdkerrors.ErrNotFound
	}
	if esc.Status != types.EscrowStatus_ESCROW_STATUS_PENDING {
		return nil, sdkerrors.ErrUnauthorized // or ErrInvalidRequest
	}
	// TODO: add sender auth / expiry checks as you need.

	modAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
	to, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}
	if err := k.bankKeeper.SendCoins(ctx, modAddr, to, sdk.NewCoins(esc.Amount)); err != nil {
		return nil, err
	}

	esc.Status = types.EscrowStatus_ESCROW_STATUS_CANCELED
	k.SetEscrow(ctx, esc) // re-write primary + index

	return &types.MsgCancelEscrowResponse{}, nil
}

// MarkEscrowClaimed sets the escrow status to CLAIMED by escrow_id
func (k Keeper) MarkEscrowClaimed(goCtx context.Context, msg *types.MsgMarkEscrowClaimed) (*types.MsgMarkEscrowClaimedResponse, error) {
    ctx := sdk.UnwrapSDKContext(goCtx)

    // Load escrow by ID
    esc, ok := k.GetEscrowByID(ctx, msg.EscrowId)
    if !ok {
        return nil, sdkerrors.ErrNotFound
    }
    if esc.Status != types.EscrowStatus_ESCROW_STATUS_PENDING {
        return nil, sdkerrors.ErrUnauthorized
    }

    // Optional: basic authorization check could be added here.
    // For now, just mark as claimed.
    esc.Status = types.EscrowStatus_ESCROW_STATUS_CLAIMED
    k.SetEscrow(ctx, esc)

    return &types.MsgMarkEscrowClaimedResponse{}, nil
}
