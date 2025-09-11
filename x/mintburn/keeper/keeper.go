package keeper

import (
	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/codec"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	mintburntypes "github.com/maany-xyz/maany-provider/x/mintburn/types"
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
    cdc            codec.BinaryCodec              
    ModuleName       string
    StoreKey         storetypes.StoreKey
    accountKeeper    mintburntypes.AccountKeeper
    bankKeeper       mintburntypes.BankKeeper
    ChannelKeeper    ChannelKeeper
	ConnectionKeeper ConnectionKeeper
	ClientKeeper     ClientKeeper

}

func NewKeeper(	cdc codec.BinaryCodec, moduleName string, storeKey storetypes.StoreKey, accountKeeper mintburntypes.AccountKeeper, bankKeeper mintburntypes.BankKeeper, channelKeeper ChannelKeeper,
	connectionKeeper ConnectionKeeper,
	clientKeeper ClientKeeper ) Keeper {
    return Keeper{
        cdc:              cdc,
        ModuleName: moduleName,
        StoreKey:   storeKey,
        accountKeeper: accountKeeper,
        bankKeeper: bankKeeper,
        ChannelKeeper:    channelKeeper,
		ConnectionKeeper: connectionKeeper,
		ClientKeeper:     clientKeeper,
    }
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/-"+"mintburn")
}

// Provider keeper
func (k Keeper) SendFromEscrowToAccount(ctx sdk.Context, escrow sdk.AccAddress, to sdk.AccAddress, coin sdk.Coin) error {
    // Move locked coins out of the ICS-20 escrow to the recipient
    if err := k.bankKeeper.SendCoins(ctx, escrow, to, sdk.NewCoins(coin)); err != nil {
        return err
    }
    return nil
}

func (k Keeper) IsAllowedChannel(ctx sdk.Context, channelID string) bool {
	store := prefix.NewStore(ctx.KVStore(k.StoreKey), []byte("allowed-channel/"))
	return store.Has([]byte(channelID))
}

// Write one escrow (overwrites if exists)
func (k Keeper) SetEscrow(ctx sdk.Context, e mintburntypes.Escrow) {
	store := ctx.KVStore(k.StoreKey)
	bz := k.cdc.MustMarshal(&e)                     
	store.Set(mintburntypes.EscrowKey(e.ConsumerChainId, e.Amount.Denom), bz)
}

// Read one escrow
func (k Keeper) GetEscrow(ctx sdk.Context, consumerID, denom string) (mintburntypes.Escrow, bool) {
	store := ctx.KVStore(k.StoreKey)
	bz := store.Get(mintburntypes.EscrowKey(consumerID, denom))
	if bz == nil {
		return mintburntypes.Escrow{}, false
	}
	var e mintburntypes.Escrow
	k.cdc.MustUnmarshal(bz, &e)                     
	return e, true
}

// iterate all escrows (useful for queries)
func (k Keeper) IterateEscrows(ctx sdk.Context, cb func(e mintburntypes.Escrow) (stop bool)) {
	ps := prefix.NewStore(ctx.KVStore(k.StoreKey), []byte(mintburntypes.EscrowPrefix+"/"))
	it := ps.Iterator(nil, nil)
	defer it.Close()

	for ; it.Valid(); it.Next() {
		var e mintburntypes.Escrow
		k.cdc.MustUnmarshal(it.Value(), &e)
		if cb(e) {
			return
		}
	}
}