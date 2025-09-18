package blockrewards

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/cosmos/cosmos-sdk/client"
    "github.com/cosmos/cosmos-sdk/types/module"
    "github.com/grpc-ecosystem/grpc-gateway/runtime"
    "github.com/spf13/cobra"

    "cosmossdk.io/core/appmodule"
    autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
    abci "github.com/cometbft/cometbft/abci/types"
    "github.com/cosmos/cosmos-sdk/codec"
    cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
    keeper "github.com/maany-xyz/maany-provider/x/blockrewards/keeper"
    "github.com/maany-xyz/maany-provider/x/blockrewards/types"
)

var (
     _ module.HasABCIGenesis = (*AppModule)(nil)
	 _ appmodule.HasEndBlocker = AppModule{}
)

// AppModuleBasic definesddd the basic application module used by the blockrewards module.
type AppModuleBasic struct{}

// Name returns the blockrewards module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the blockrewards module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the blockrewards module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {}

// GetTxCmd returns the root tx command for the blockrewards module.
func (AppModuleBasic) GetTxCmd() *cobra.Command { return nil }

// GetQueryCmd returns the root query command for the blockrewards module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command { return nil }

// AppModule implements the AppModule interface for the blockrewards module.
type AppModule struct {
	cdc codec.Codec
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(cdc codec.Codec, k keeper.Keeper) AppModule {
	return AppModule{
		cdc: cdc,
		AppModuleBasic: AppModuleBasic{},
		keeper:         k,
	}
}

// IsAppModule is a marker method to identify AppModules
func (AppModule) IsAppModule() {}
func (AppModule) IsOnePerModuleType() {}
// RegisterInvariants registers the invariants for the blockrewards module.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// RegisterInterfaces registers the module's protobuf interfaces.
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// QuerierRoute returns the blockrewards module's query routing key.
func (AppModule) QuerierRoute() string { return types.ModuleName }

// RegisterServices registers the module's services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
    types.RegisterMsgServer(cfg.MsgServer(), am.keeper)
    types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

func (am AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
    // Return the default genesis state marshaled as JSON
    gen := types.DefaultGenesisState()
    return cdc.MustMarshalJSON(&gen)
}

func (am AppModule) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
    // Unmarshal the genesis state from the JSON
    var genState types.GenesisState
    if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
        return fmt.Errorf("failed to unmarshal genesis state: %w", err)
    }

    // Validate the genesis state
    return genState.Validate()
}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
    var genesisState types.GenesisState
    cdc.MustUnmarshalJSON(data, &genesisState)
    am.keeper.InitGenesis(ctx, &genesisState)
    return nil
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
     gs := am.keeper.ExportGenesis(ctx)
    return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock executes all logic for the blockrewards module at the beginning of a block.
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestFinalizeBlock) {}

// EndBlock executes all logic for the blockrewards module at the end of a block.
func (am AppModule) EndBlock(ctx context.Context) error {
    sdkCtx := sdk.UnwrapSDKContext(ctx)
    sdkCtx.Logger().Info("Executing EndBlock for blockrewards module")
    EndBlocker(sdkCtx, ctx, am.keeper)
    return nil
}
// func (am AppModule) EndBlock(context Conte) (error) {
// 	ctx.Logger().Info("Entering EndBlock for blockrewards module")
// 	EndBlocker(ctx, am.keeper)
// 	return nil
// }
// EndBlocker is the core logic for the blockrewards module at the end of each block.
func EndBlocker (sdkContext sdk.Context, ctx context.Context, k keeper.Keeper) {
    // Retrieve runtime params (governable via x/params)
    mp := k.GetModuleParams(sdkContext)

    // Determine effective reward with optional halving schedule
    amt := mp.RewardAmount.Amount
    if mp.HalvingInterval > 0 {
        height := uint64(sdkContext.BlockHeight())
        halvings := height / mp.HalvingInterval
        for i := uint64(0); i < halvings && !amt.IsZero(); i++ {
            // integer halving
            amt = amt.QuoRaw(2)
        }
    }

    if amt.IsZero() {
        // nothing to distribute
        return
    }

    coin := sdk.NewCoin(mp.RewardAmount.Denom, amt)
    reward := sdk.NewCoins(coin)

    // Attempt distribution; if module account lacks funds, we'll log and stop.
    if err := k.DistributeRewards(sdkContext, ctx, reward); err != nil {
        sdkContext.Logger().Error("blockrewards: distribution failed", "err", err)
    }
}

// AutoCLIOptions returns CLI wiring for v0.50 autocli
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
    return &autocliv1.ModuleOptions{
        Query: &autocliv1.ServiceCommandDescriptor{
            Service: "maany.blockrewards.v1.Query",
            RpcCommandOptions: []*autocliv1.RpcCommandOptions{
                {
                    RpcMethod: "Params",
                    Use:       "params",
                    Short:     "Show blockrewards module params",
                },
                {
                    RpcMethod: "EffectiveReward",
                    Use:       "effective-reward",
                    Short:     "Compute effective reward at a height",
                    FlagOptions: map[string]*autocliv1.FlagOptions{
                        "height": {Name: "at-height", Usage: "Block height to evaluate at (default: current)"},
                    },
                },
            },
        },
    }
}
