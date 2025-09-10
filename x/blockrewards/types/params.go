package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultParams returns the default parameters for the blockrewards module.
func DefaultParams() Params {
	return Params{
		BlockRewardAmount: sdk.NewCoin("stake", math.NewInt(100000)),
	}
}


// Validate performs validation on the blockrewards parameters.
func (p Params) Validate() error {
	if p.BlockRewardAmount.IsNegative() {
		return fmt.Errorf("block reward amount cannot be negative: %s", p.BlockRewardAmount)
	}
	return nil
}
