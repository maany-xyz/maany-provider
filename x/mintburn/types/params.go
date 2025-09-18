package types

import (
    "fmt"

    paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keys for mintburn module parameters
var (
    KeyAllowedConsumers = []byte("AllowedConsumers")
    KeyRejectUnauthorizedChannels = []byte("RejectUnauthorizedChannels")
    KeyAllowedClientIDs = []byte("AllowedClientIDs")
    KeyAllowedConnectionIDs = []byte("AllowedConnectionIDs")
    KeyAllowedFunders = []byte("AllowedFunders")
)

// DefaultParams returns the default params for the generated Params type
func DefaultParams() Params {
    return Params{AllowedConsumers: []string{}}
}

// Validate performs basic validation of params
func (p Params) Validate() error {
    for _, id := range p.AllowedConsumers {
        if len(id) == 0 {
            return fmt.Errorf("allowed consumer chain-id must not be empty")
        }
    }
    return nil
}

// ParamKeyTable for mintburn module
func ParamKeyTable() paramstypes.KeyTable {
    return paramstypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface on the generated Params type
func (p *Params) ParamSetPairs() paramstypes.ParamSetPairs {
    return paramstypes.ParamSetPairs{
        paramstypes.NewParamSetPair(KeyAllowedConsumers, &p.AllowedConsumers, validateAllowedConsumers),
        paramstypes.NewParamSetPair(KeyRejectUnauthorizedChannels, &p.RejectUnauthorizedChannels, validateBool),
        paramstypes.NewParamSetPair(KeyAllowedClientIDs, &p.AllowedClientIds, validateStringSlice),
        paramstypes.NewParamSetPair(KeyAllowedConnectionIDs, &p.AllowedConnectionIds, validateStringSlice),
        paramstypes.NewParamSetPair(KeyAllowedFunders, &p.AllowedFunders, validateStringSlice),
    }
}

func validateAllowedConsumers(i interface{}) error {
    v, ok := i.([]string)
    if !ok {
        return fmt.Errorf("invalid type for AllowedConsumers: %T", i)
    }
    for _, id := range v {
        if len(id) == 0 {
            return fmt.Errorf("allowed consumer chain-id must not be empty")
        }
    }
    return nil
}

func validateStringSlice(i interface{}) error {
    v, ok := i.([]string)
    if !ok {
        return fmt.Errorf("invalid type for []string: %T", i)
    }
    for _, s := range v {
        if len(s) == 0 {
            return fmt.Errorf("string value must not be empty")
        }
    }
    return nil
}

func validateBool(i interface{}) error {
    if _, ok := i.(bool); !ok {
        return fmt.Errorf("invalid type for bool: %T", i)
    }
    return nil
}
