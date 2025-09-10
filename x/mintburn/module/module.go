package mintburn

import (
	"encoding/json"

	"cosmossdk.io/log"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	keeper "github.com/maany-xyz/maany-provider/x/mintburn/keeper"
	mintburntypes "github.com/maany-xyz/maany-provider/x/mintburn/types"
)

var (
    _ module.AppModule      = (*AppModule)(nil)
    _ module.AppModuleBasic = (*AppModuleBasic)(nil)
    _ module.HasABCIGenesis = (*AppModule)(nil)
)

// AppModuleBasic defines the basic application module used by the mintburn module.
type AppModuleBasic struct{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func (am AppModule) IsAppModule() {}

// Name returns the mintburn module's name.
func (AppModuleBasic) Name() string {
    return mintburntypes.ModuleName
}

// RegisterLegacyAminoCodec registers the mintburn module's types for Amino.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers the mintburn module's interface types.
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {}

// DefaultGenesis returns the mintburn module's default genesis state as raw bytes.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
    return nil // Provide default genesis state if needed
}

// ValidateGenesis validates the mintburn module's genesis state.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
    return nil
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// err := providertypes.RegisterQueryHandlerClient(context.Background(), mux, providertypes.NewQueryClient(clientCtx))
	// if err != nil {
	// 	// same behavior as in cosmos-sdk
	// 	panic(err)
	// }
}

// AppModule implements the AppModule interface for the mintburn module.
type AppModule struct {
	cdc codec.Codec
    AppModuleBasic
    keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(cdc codec.Codec, k keeper.Keeper, appLogger log.Logger) AppModule {
    return AppModule{
		cdc: cdc,
		AppModuleBasic: AppModuleBasic{},
        keeper: k,
    }
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports the genesis state for the mintburn module.
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return nil
}