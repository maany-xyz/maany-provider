package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error

    SendCoinsFromModuleToAccount(ctx context.Context, moduleName string, recipient sdk.AccAddress, amt sdk.Coins) error
    SendCoinsFromAccountToModule(ctx context.Context, sender sdk.AccAddress, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error

	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, module string, amt sdk.Coins) error

    HasBalance(ctx context.Context, addr sdk.AccAddress, amt sdk.Coin) bool
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}


type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
}