package types

import fmt "fmt"

// Validate checks the validity of the GenesisState.
func (gs GenesisState) Validate() error {

    if err := gs.Params.Validate(); err != nil {
        return fmt.Errorf("invalid params: %w", err)
    }
    return nil
}