package types

import (
    "fmt"

    "cosmossdk.io/math"
    sdk "github.com/cosmos/cosmos-sdk/types"
    paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Runtime/governable params stored in x/params Subspace
type ModuleParams struct {
    RewardAmount    sdk.Coin `json:"reward_amount"`
    // HalvingInterval in blocks; if 0, halving is disabled
    HalvingInterval uint64   `json:"halving_interval"`
}

// Param keys
var (
    KeyRewardAmount    = []byte("RewardAmount")
    KeyHalvingInterval = []byte("HalvingInterval")
)

func DefaultModuleParams() ModuleParams {
    return ModuleParams{
        // default denom should align with bond denom; this is only a fallback
        RewardAmount:    sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100000)),
        HalvingInterval: 0, // disabled by default; can be enabled via gov
    }
}

// Validate checks param values
func (p ModuleParams) Validate() error {
    if p.RewardAmount.Amount.IsNegative() {
        return fmt.Errorf("reward amount cannot be negative: %s", p.RewardAmount)
    }
    if p.RewardAmount.Denom == "" {
        return fmt.Errorf("reward denom cannot be empty")
    }
    // HalvingInterval can be zero to disable
    return nil
}

// ParamSetPairs implements ParamSet
func (p *ModuleParams) ParamSetPairs() paramstypes.ParamSetPairs {
    return paramstypes.ParamSetPairs{
        paramstypes.NewParamSetPair(KeyRewardAmount, &p.RewardAmount, validateCoinNonNegative),
        paramstypes.NewParamSetPair(KeyHalvingInterval, &p.HalvingInterval, validateUint64),
    }
}

func ParamKeyTable() paramstypes.KeyTable { return paramstypes.NewKeyTable().RegisterParamSet(&ModuleParams{}) }

// --- validators ---
func validateCoinNonNegative(i interface{}) error {
    c, ok := i.(sdk.Coin)
    if !ok {
        return fmt.Errorf("invalid type for coin: %T", i)
    }
    if c.Denom == "" {
        return fmt.Errorf("coin denom cannot be empty")
    }
    if c.Amount.IsNegative() {
        return fmt.Errorf("coin amount cannot be negative")
    }
    return nil
}

func validateUint64(i interface{}) error {
    _, ok := i.(uint64)
    if !ok {
        return fmt.Errorf("invalid type for uint64: %T", i)
    }
    return nil
}

