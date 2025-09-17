package keepers

import (
    sdk "github.com/cosmos/cosmos-sdk/types"

    icsproviderkeeper "github.com/cosmos/interchain-security/v5/x/ccv/provider/keeper"

    mintburntypes "github.com/maany-xyz/maany-provider/x/mintburn/types"
)

// Ensure ProviderRegistryAdapter implements the mintburn ProviderConsumerRegistry interface
var _ mintburntypes.ProviderConsumerRegistry = (*ProviderRegistryAdapter)(nil)

// ProviderRegistryAdapter wraps the CCV provider keeper to expose a minimal
// interface needed by the mintburn module for channel authorization.
type ProviderRegistryAdapter struct {
    prov icsproviderkeeper.Keeper
}

func NewProviderRegistryAdapter(prov icsproviderkeeper.Keeper) *ProviderRegistryAdapter {
    return &ProviderRegistryAdapter{prov: prov}
}

// IsConsumerClientRegistered returns true if the provider keeper records the
// given consumer chain-id as registered and its associated local client id
// matches the provided clientID.
func (a *ProviderRegistryAdapter) IsConsumerClientRegistered(ctx sdk.Context, clientID string, consumerChainID string) bool {
    // Note: exact API may differ across ICS versions. The intended logic is:
    //   storedClientID, found := a.prov.GetConsumerClientId(ctx, consumerChainID)
    //   return found && storedClientID == clientID
    // If the provider keeper exposes a different method name, adapt here.

    storedClientID, found := a.prov.GetConsumerClientId(ctx, consumerChainID)
    if !found {
        return false
    }
    return storedClientID == clientID
}
