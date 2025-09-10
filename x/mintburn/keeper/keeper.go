package mintburn

import (
	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)


type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChannel string) (channeltypes.Channel, bool)
}

type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
}

type ClientKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (ibcexported.ClientState, bool)
}


type Keeper struct {
    ModuleName string
    StoreKey      storetypes.StoreKey
    BankKeeper bankKeeper.Keeper
    ChannelKeeper    ChannelKeeper
	ConnectionKeeper ConnectionKeeper
	ClientKeeper     ClientKeeper

}

func NewKeeper(moduleName string, storeKey storetypes.StoreKey, bankKeeper bankKeeper.Keeper, channelKeeper ChannelKeeper,
	connectionKeeper ConnectionKeeper,
	clientKeeper ClientKeeper ) Keeper {
    return Keeper{
        ModuleName: moduleName,
        StoreKey:   storeKey,
        BankKeeper: bankKeeper,
        ChannelKeeper:    channelKeeper,
		ConnectionKeeper: connectionKeeper,
		ClientKeeper:     clientKeeper,
    }
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/-"+"mintburn")
}

func (k Keeper) MintTokens(ctx sdk.Context, recipient sdk.AccAddress, amount sdk.Coin) error {
    // Mint tokens to the module account
    ctx.Logger().Info("Now in MintTokens in Keeper")
    err := k.BankKeeper.MintCoins(ctx, k.ModuleName, sdk.NewCoins(amount))
    if err != nil {
        return err
    }

    // Send minted tokens to the recipient
    return k.BankKeeper.SendCoinsFromModuleToAccount(ctx, k.ModuleName, recipient, sdk.NewCoins(amount))
}

func (k Keeper) GetBalances(ctx sdk.Context, address sdk.AccAddress) error {
    ctx.Logger().Info("Now in Get Balance in Keeper")
    balances := k.BankKeeper.GetAllBalances(ctx, address)
    ctx.Logger().Info("The transfer module account balances are: ", "balamces", balances.String())
    return nil
}

func (k Keeper) BurnTokens(ctx sdk.Context, recipient sdk.AccAddress, amount sdk.Coin) error {

    if err := k.BankKeeper.SendCoinsFromModuleToModule(ctx, ibctransfertypes.ModuleName, k.ModuleName, sdk.NewCoins(amount)); err != nil {
		ctx.Logger().Error("Failed to redirect tokens to module account", "error", err)
	    return err
    }

    return nil
}

// Provider keeper
func (k Keeper) SendFromEscrowToAccount(ctx sdk.Context, escrow sdk.AccAddress, to sdk.AccAddress, coin sdk.Coin) error {
    // Move locked coins out of the ICS-20 escrow to the recipient
    if err := k.BankKeeper.SendCoins(ctx, escrow, to, sdk.NewCoins(coin)); err != nil {
        return err
    }
    return nil
}

func (k Keeper) IsAllowedChannel(ctx sdk.Context, channelID string) bool {
	store := prefix.NewStore(ctx.KVStore(k.StoreKey), []byte("allowed-channel/"))
	return store.Has([]byte(channelID))
}

func (k Keeper) BurnEscrowedTokens(ctx sdk.Context, escrowAddr sdk.AccAddress, coin sdk.Coin) error {
    	if err := k.BankKeeper.SendCoinsFromAccountToModule(ctx, escrowAddr, k.ModuleName, sdk.NewCoins(coin)); err != nil {
		return err
	}

	return k.BankKeeper.BurnCoins(ctx, k.ModuleName, sdk.NewCoins(coin))
}