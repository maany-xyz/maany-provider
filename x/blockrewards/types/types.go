package types

import (
	"cosmossdk.io/math"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new genesis state with default values.
func NewGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// ModuleName defines the name of the module
// DefaultGenesisState returns the default genesis state for the blockrewards module.
func DefaultGenesisState() GenesisState {
    return GenesisState{
		Params: Params{
			BlockRewardAmount:  sdk.NewCoin("stake", math.NewInt(100000)) ,
		},
    }
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	// If you have interfaces to register, add them here. For example:
	// registry.RegisterImplementations(
	//     (*sdk.Msg)(nil), // Register any Msg types here if needed
	// )
}

