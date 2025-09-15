package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ValidateBasic for MsgEscrowInitial
func (m *MsgEscrowInitial) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "sender: %v", err)
	}
	if strings.TrimSpace(m.ConsumerChainId) == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "consumer_chain_id is required")
	}
	if !m.Amount.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "invalid amount")
	}
	if !m.Amount.Amount.IsPositive() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, "amount must be > 0")
	}
	if err := sdk.ValidateDenom(m.Amount.Denom); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "denom: %v", err)
	}
	if m.Recipient != "" {
		if _, err := sdk.AccAddressFromBech32(m.Recipient); err != nil {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "recipient: %v", err)
		}
	}
	return nil
}

// ValidateBasic for MsgCancelEscrow
func (m *MsgCancelEscrow) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "sender: %v", err)
	}
	if strings.TrimSpace(m.ConsumerChainId) == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "consumer_chain_id is required")
	}
	if err := sdk.ValidateDenom(m.Denom); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "denom: %v", err)
	}
	return nil
}

// ValidateBasic for MsgMarkEscrowClaimed
func (m *MsgMarkEscrowClaimed) ValidateBasic() error {
    if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
        return errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "sender: %v", err)
    }
    if strings.TrimSpace(m.EscrowId) == "" {
        return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "escrow_id is required")
    }
    if strings.TrimSpace(m.ConsumerChainId) == "" {
        return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "consumer_chain_id is required")
    }
    return nil
}
