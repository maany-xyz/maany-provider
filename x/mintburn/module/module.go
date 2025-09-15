package mintburn

import (
	"encoding/json"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/log"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	cli "github.com/maany-xyz/maany-provider/x/mintburn/client/cli"
	keeper "github.com/maany-xyz/maany-provider/x/mintburn/keeper"
	mintburntypes "github.com/maany-xyz/maany-provider/x/mintburn/types"
	"github.com/spf13/cobra"
)

var (
	_ module.AppModule      = (*AppModule)(nil)
	_ module.AppModuleBasic = (*AppModuleBasic)(nil)
	_ module.HasABCIGenesis = (*AppModule)(nil)
)

func (AppModuleBasic) GetTxCmd() *cobra.Command    { return cli.NewTxCmd() }
func (AppModuleBasic) GetQueryCmd() *cobra.Command { return nil } 

// -----------------------------
// AppModuleBasic
// -----------------------------

type AppModuleBasic struct{}

func (am AppModule) IsOnePerModuleType() {}
func (am AppModule) IsAppModule()        {}

func (AppModuleBasic) Name() string { return mintburntypes.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

func (AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	mintburntypes.RegisterInterfaces(reg)
}

func (AppModuleBasic) DefaultGenesis(codec.JSONCodec) json.RawMessage {
	return []byte(`{}`)
}

func (AppModuleBasic) ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error {
	return nil
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// If/when you generate REST handlers, register them here with:
	// _ = mintburntypes.RegisterQueryHandlerClient(context.Background(), mux, mintburntypes.NewQueryClient(clientCtx))
}

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
  return &autocliv1.ModuleOptions{
    Query: &autocliv1.ServiceCommandDescriptor{
      Service: "maany.mintburn.v1.Query",
      RpcCommandOptions: []*autocliv1.RpcCommandOptions{
        {
          RpcMethod: "Escrow",
          Use:       "escrow [consumer-chain-id] [denom]",
          Short:     "Query a single escrow",
          PositionalArgs: []*autocliv1.PositionalArgDescriptor{
            {ProtoField: "consumer_chain_id"},
            {ProtoField: "denom"},
          },
        },
        {
          RpcMethod: "Escrows",
          Use:       "escrows",
          Short:     "List all escrows",
          FlagOptions: map[string]*autocliv1.FlagOptions{
            // key = proto field name
            "status_filter": {Name: "status-filter", Usage: "PENDING|CLAIMED|CANCELED"},
          },
        },
        {
          RpcMethod: "EscrowProof",
          Use:       "escrow-proof [consumer-chain-id] [denom]",
          Short:     "Export ICS-23 proof bundle for an escrow",
          PositionalArgs: []*autocliv1.PositionalArgDescriptor{
            {ProtoField: "consumer_chain_id"},
            {ProtoField: "denom"},
          },
          // Avoid collision with global --height by renaming the flag
          FlagOptions: map[string]*autocliv1.FlagOptions{
            "height": {Name: "prove-height", Usage: "Provider block height to prove"},
          },
        },
      },
    },
  }
}

// -----------------------------
// AppModule
// -----------------------------

type AppModule struct {
	cdc codec.Codec
	keeper keeper.Keeper
	AppModuleBasic
}

func NewAppModule(cdc codec.Codec, k keeper.Keeper, _ log.Logger) AppModule {
	return AppModule{
		cdc:            cdc,
		keeper:         k,
		AppModuleBasic: AppModuleBasic{},
	}
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
    mintburntypes.RegisterMsgServer(cfg.MsgServer(), am.keeper)
    mintburntypes.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(am.keeper))
}

func (am AppModule) InitGenesis(sdk.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate {
	return nil
}

func (am AppModule) ExportGenesis(sdk.Context, codec.JSONCodec) json.RawMessage {
	return []byte(`{}`)
}

func (am AppModule) ConsensusVersion() uint64 { return 1 }
