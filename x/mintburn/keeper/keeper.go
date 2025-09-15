package keeper

import (
	"cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	mintburntypes "github.com/maany-xyz/maany-provider/x/mintburn/types"
)

// ---- IBC keeper interfaces (unchanged) ----
type ChannelKeeper interface {
	GetChannel(ctx sdk.Context, srcPort, srcChannel string) (channeltypes.Channel, bool)
}

type ConnectionKeeper interface {
	GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool)
}

type ClientKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (ibcexported.ClientState, bool)
}

// ---- Keeper ----
type Keeper struct {
	cdc              codec.BinaryCodec
	ModuleName       string
	StoreKey         storetypes.StoreKey
	accountKeeper    mintburntypes.AccountKeeper
	bankKeeper       mintburntypes.BankKeeper
	ChannelKeeper    ChannelKeeper
	ConnectionKeeper ConnectionKeeper
	ClientKeeper     ClientKeeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	moduleName string,
	storeKey storetypes.StoreKey,
	accountKeeper mintburntypes.AccountKeeper,
	bankKeeper mintburntypes.BankKeeper,
	channelKeeper ChannelKeeper,
	connectionKeeper ConnectionKeeper,
	clientKeeper ClientKeeper,
) Keeper {
	return Keeper{
		cdc:              cdc,
		ModuleName:       moduleName,
		StoreKey:         storeKey,
		accountKeeper:    accountKeeper,
		bankKeeper:       bankKeeper,
		ChannelKeeper:    channelKeeper,
		ConnectionKeeper: connectionKeeper,
		ClientKeeper:     clientKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+k.ModuleName)
}

// Provider keeper
func (k Keeper) SendFromEscrowToAccount(ctx sdk.Context, escrow sdk.AccAddress, to sdk.AccAddress, coin sdk.Coin) error {
	return k.bankKeeper.SendCoins(ctx, escrow, to, sdk.NewCoins(coin))
}

func (k Keeper) IsAllowedChannel(ctx sdk.Context, channelID string) bool {
	ps := prefix.NewStore(ctx.KVStore(k.StoreKey), []byte("allowed-channel/"))
	return ps.Has([]byte(channelID))
}

// =========================
// Escrow storage helpers
// Primary:    EscrowPrefix + escrow_id -> Escrow (value)
// Secondary:  EscrowIndexPrefix + consumer_chain_id + 0x00 + denom -> escrow_id (value)
// =========================

// Write one escrow (overwrites if exists) and maintain index
func (k Keeper) SetEscrow(ctx sdk.Context, e mintburntypes.Escrow) {
	store := ctx.KVStore(k.StoreKey)

	// 1) primary
	bz := k.cdc.MustMarshal(&e)
	store.Set(mintburntypes.EscrowKeyByID(e.EscrowId), bz)

	// 2) secondary index (consumer_chain_id, denom) -> escrow_id
	idx := mintburntypes.EscrowIndexKey(e.ConsumerChainId, e.Amount.Denom)
	store.Set(idx, []byte(e.EscrowId))
}

// Read one escrow by (consumer_chain_id, denom) using the secondary index
func (k Keeper) GetEscrow(ctx sdk.Context, consumerID, denom string) (mintburntypes.Escrow, bool) {
	store := ctx.KVStore(k.StoreKey)

	// 1) resolve escrow_id from index
	idbz := store.Get(mintburntypes.EscrowIndexKey(consumerID, denom))
	if idbz == nil {
		return mintburntypes.Escrow{}, false
	}
	escrowID := string(idbz)

	// 2) fetch primary record
	bz := store.Get(mintburntypes.EscrowKeyByID(escrowID))
	if bz == nil {
		return mintburntypes.Escrow{}, false
	}

	var e mintburntypes.Escrow
	k.cdc.MustUnmarshal(bz, &e)
	// sanity: ensure EscrowId present
	if e.EscrowId == "" {
		e.EscrowId = escrowID
	}
	return e, true
}

// Iterate all escrows (useful for queries)
func (k Keeper) IterateEscrows(ctx sdk.Context, cb func(e mintburntypes.Escrow) (stop bool)) {
	// Use a prefix store over the primary prefix bytes (DO NOT append a "/" string).
	ps := prefix.NewStore(ctx.KVStore(k.StoreKey), mintburntypes.EscrowPrefix)
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
