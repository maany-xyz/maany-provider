package types

import (
    "cosmossdk.io/math"
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
            BlockRewardAmount:  sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000)) ,
        },
    }
}

// Interface registrations are defined in codec.go
