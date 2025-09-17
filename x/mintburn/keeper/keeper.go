package keeper

import (
    "cosmossdk.io/log"
    "cosmossdk.io/store/prefix"
    storetypes "cosmossdk.io/store/types"
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

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
    paramSpace       paramstypes.Subspace
    accountKeeper    mintburntypes.AccountKeeper
    bankKeeper       mintburntypes.BankKeeper
    ChannelKeeper    ChannelKeeper
    ConnectionKeeper ConnectionKeeper
    ClientKeeper     ClientKeeper
    // Optional: CCV provider consumer registry for channel auth
    providerRegistry mintburntypes.ProviderConsumerRegistry
}

func NewKeeper(
    cdc codec.BinaryCodec,
    moduleName string,
    storeKey storetypes.StoreKey,
    paramSpace paramstypes.Subspace,
    accountKeeper mintburntypes.AccountKeeper,
    bankKeeper mintburntypes.BankKeeper,
    channelKeeper ChannelKeeper,
    connectionKeeper ConnectionKeeper,
    clientKeeper ClientKeeper,
    providerRegistry mintburntypes.ProviderConsumerRegistry,
) Keeper {
    return Keeper{
        cdc:              cdc,
        ModuleName:       moduleName,
        StoreKey:         storeKey,
        paramSpace:       paramSpace,
        accountKeeper:    accountKeeper,
        bankKeeper:       bankKeeper,
        ChannelKeeper:    channelKeeper,
        ConnectionKeeper: connectionKeeper,
        ClientKeeper:     clientKeeper,
        providerRegistry: providerRegistry,
    }
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
    return ctx.Logger().With("module", "x/"+k.ModuleName)
}

// ProviderRegistry exposes the optional provider registry
func (k Keeper) ProviderRegistry() mintburntypes.ProviderConsumerRegistry {
    return k.providerRegistry
}

// -----------------
// Params helpers
// -----------------

func (k Keeper) GetParams(ctx sdk.Context) mintburntypes.Params {
    // If subspace not present or key not set, return defaults
    if k.paramSpace.Name() == "" || !k.paramSpace.Has(ctx, mintburntypes.KeyAllowedConsumers) {
        return mintburntypes.DefaultParams()
    }
    var p mintburntypes.Params
    k.paramSpace.GetParamSet(ctx, &p)
    if err := p.Validate(); err != nil {
        return mintburntypes.DefaultParams()
    }
    return p
}

func (k Keeper) IsAllowedConsumerByParams(ctx sdk.Context, chainID string) bool {
    p := k.GetParams(ctx)
    for _, id := range p.AllowedConsumers {
        if id == chainID {
            return true
        }
    }
    return false
}

// SetParams sets the module parameters.
func (k Keeper) SetParams(ctx sdk.Context, p mintburntypes.Params) {
    if k.paramSpace.Name() == "" {
        return
    }
    k.paramSpace.SetParamSet(ctx, &p)
}

// InitGenesis sets initial state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, gs *mintburntypes.GenesisState) {
    if gs == nil {
        gs = &mintburntypes.GenesisState{Params: &mintburntypes.Params{}}
    }
    if gs.Params == nil {
        k.SetParams(ctx, mintburntypes.Params{})
    } else {
        k.SetParams(ctx, *gs.Params)
    }
}

// ExportGenesis returns the current state as genesis
func (k Keeper) ExportGenesis(ctx sdk.Context) *mintburntypes.GenesisState {
    p := k.GetParams(ctx)
    return &mintburntypes.GenesisState{Params: &p}
}

// Provider keeper
func (k Keeper) SendFromEscrowToAccount(ctx sdk.Context, escrow sdk.AccAddress, to sdk.AccAddress, coin sdk.Coin) error {
	return k.bankKeeper.SendCoins(ctx, escrow, to, sdk.NewCoins(coin))
}

func (k Keeper) IsAllowedChannel(ctx sdk.Context, channelID string) bool {
    ps := prefix.NewStore(ctx.KVStore(k.StoreKey), []byte("allowed-channel/"))
    return ps.Has([]byte(channelID))
}

// Maintain a mapping from consumer chain-id -> set of allowed transfer channel IDs
func (k Keeper) AddAllowedChannelForChain(ctx sdk.Context, chainID, channelID string) {
    ps := prefix.NewStore(ctx.KVStore(k.StoreKey), []byte("allowed-channel-by-chain/"))
    // key layout: "allowed-channel-by-chain/" + chainID + "/" + channelID
    key := append([]byte(chainID), '/')
    key = append(key, []byte(channelID)...)
    ps.Set(key, []byte{1})
}

func (k Keeper) RemoveAllowedChannelForChain(ctx sdk.Context, chainID, channelID string) {
    ps := prefix.NewStore(ctx.KVStore(k.StoreKey), []byte("allowed-channel-by-chain/"))
    key := append([]byte(chainID), '/')
    key = append(key, []byte(channelID)...)
    ps.Delete(key)
}

// GetAnyAllowedChannelForChain returns one allowed channel for a consumer chain-id if present
func (k Keeper) GetAnyAllowedChannelForChain(ctx sdk.Context, chainID string) (string, bool) {
    ps := prefix.NewStore(ctx.KVStore(k.StoreKey), []byte("allowed-channel-by-chain/"))
    // iterate keys starting with chainID + "/"
    pfx := append([]byte(chainID), '/')
    it := ps.Iterator(pfx, nil)
    defer it.Close()
    if it.Valid() {
        // extract channelID from key suffix
        key := it.Key()
        // key format: chainID + '/' + channelID
        // find last '/'
        lastSlash := -1
        for i := len(key) - 1; i >= 0; i-- {
            if key[i] == '/' {
                lastSlash = i
                break
            }
        }
        if lastSlash >= 0 && lastSlash+1 < len(key) {
            return string(key[lastSlash+1:]), true
        }
    }
    return "", false
}

// (canonical channel helpers intentionally omitted; multiple channels allowed for now)

// (KV allow-list of consumer chain-ids removed; use params + provider registry)

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

// GetEscrowByID fetches an escrow by its primary id key
func (k Keeper) GetEscrowByID(ctx sdk.Context, escrowID string) (mintburntypes.Escrow, bool) {
    store := ctx.KVStore(k.StoreKey)
    bz := store.Get(mintburntypes.EscrowKeyByID(escrowID))
    if bz == nil {
        return mintburntypes.Escrow{}, false
    }
    var e mintburntypes.Escrow
    k.cdc.MustUnmarshal(bz, &e)
    if e.EscrowId == "" {
        e.EscrowId = escrowID
    }
    return e, true
}

// Authorized ICA mapping helpers
func (k Keeper) SetAuthorizedICA(ctx sdk.Context, consumerChainID, icaBech32 string) {
    store := ctx.KVStore(k.StoreKey)
    store.Set(mintburntypes.AuthorizedICAKey(consumerChainID), []byte(icaBech32))
}

func (k Keeper) GetAuthorizedICA(ctx sdk.Context, consumerChainID string) (string, bool) {
    store := ctx.KVStore(k.StoreKey)
    bz := store.Get(mintburntypes.AuthorizedICAKey(consumerChainID))
    if bz == nil {
        return "", false
    }
    return string(bz), true
}
