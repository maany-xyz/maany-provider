package mintburn

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgMintTokens struct {
    Sender sdk.AccAddress `json:"sender"`
    Amount sdk.Coin       `json:"amount"`
}

// MsgBurnTokens defines a message for burning tokens.
type MsgBurnTokens struct {
    Sender sdk.AccAddress `json:"sender"`
    Amount sdk.Coin       `json:"amount"`
}

type BankKeeper interface {
    MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
    BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
    SendCoinsFromModuleToAccount(ctx sdk.Context, moduleName string, recipient sdk.AccAddress, amt sdk.Coins) error
    SendCoinsFromAccountToModule(ctx sdk.Context, sender sdk.AccAddress, moduleName string, amt sdk.Coins) error
    HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool
}