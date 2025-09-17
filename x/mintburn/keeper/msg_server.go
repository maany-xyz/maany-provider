package keeper

import (
	"context"
	"encoding/binary"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
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
    // Consumer chain check must match the escrow record
    if msg.ConsumerChainId == "" || msg.ConsumerChainId != esc.ConsumerChainId {
        return nil, sdkerrors.ErrUnauthorized
    }

    // Authorization: strictly require the authorized ICA address for this consumer.
    if ica, found := k.GetAuthorizedICA(ctx, msg.ConsumerChainId); !found {
        ctx.Logger().Error("mintburn: CLAIMED rejected — no authorized ICA registered",
            "consumer_chain_id", msg.ConsumerChainId, "sender", msg.Sender)
        return nil, sdkerrors.ErrUnauthorized
    } else if msg.Sender != ica {
        ctx.Logger().Error("mintburn: CLAIMED rejected — sender is not authorized ICA",
            "consumer_chain_id", msg.ConsumerChainId, "sender", msg.Sender, "expected_ica", ica)
        return nil, sdkerrors.ErrUnauthorized
    }

    // Idempotent: already claimed -> OK
    if esc.Status == types.EscrowStatus_ESCROW_STATUS_CLAIMED {
        return &types.MsgMarkEscrowClaimedResponse{}, nil
    }
    if esc.Status != types.EscrowStatus_ESCROW_STATUS_PENDING {
        return nil, sdkerrors.ErrInvalidRequest
    }

    // Ensure at least one allowed transfer channel exists for this consumer chain-id
    if _, ok := k.GetAnyAllowedChannelForChain(ctx, esc.ConsumerChainId); !ok {
        return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "no allowed transfer channel for chain-id %s", esc.ConsumerChainId)
    }

    // Move funds from mintburn module account to the dedicated mintburn escrow address
    mintburnEscrowID := types.MintburnEscrowChannelIDFor(esc.ConsumerChainId)
    escrowAddr := ibctransfertypes.GetEscrowAddress("transfer", mintburnEscrowID)
    modAddr := k.accountKeeper.GetModuleAddress(types.ModuleName)
    amt := sdk.NewCoins(esc.Amount)

    // Log before transfer
    ctx.Logger().Info("mintburn: funding transfer escrow from module",
        "escrow_id", esc.EscrowId,
        "consumer_chain_id", esc.ConsumerChainId,
        "channel_id", mintburnEscrowID,
        "amount", esc.Amount.String(),
        "from", modAddr.String(),
        "to", escrowAddr.String(),
    )

    if err := k.bankKeeper.SendCoins(ctx, modAddr, escrowAddr, amt); err != nil {
        return nil, err
    }

    // Mark as claimed after successful movement
    esc.Status = types.EscrowStatus_ESCROW_STATUS_CLAIMED
    k.SetEscrow(ctx, esc)

    // Emit event for observability
    ctx.EventManager().EmitEvents(sdk.Events{
        sdk.NewEvent(
            "mintburn_mark_escrow_claimed",
            sdk.NewAttribute("escrow_id", esc.EscrowId),
            sdk.NewAttribute("consumer_chain_id", esc.ConsumerChainId),
            sdk.NewAttribute("denom", esc.Amount.Denom),
            sdk.NewAttribute("channel_id", mintburnEscrowID),
            sdk.NewAttribute("moved_to_escrow", esc.Amount.String()),
        ),
    })

    return &types.MsgMarkEscrowClaimedResponse{}, nil
}
