package keeper

import (
    "context"
    "encoding/hex"
    "fmt"

    storetypes "cosmossdk.io/store/types"

    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
    accountKeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
    bankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
    stakingKeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
    "github.com/maany-xyz/maany-provider/x/blockrewards/types"
)

// Keeper defines the blockrewards module's keeper
type Keeper struct {
    cdc           codec.BinaryCodec
    storeKey      storetypes.StoreKey
    bankKeeper    bankKeeper.Keeper
    stakingKeeper stakingKeeper.Keeper
    accountKeeper accountKeeper.AccountKeeper
    paramSpace    paramstypes.Subspace
    authority     string
}

// NewKeeper creates a new blockrewards Keeper instance
func NewKeeper (
    cdc           codec.BinaryCodec,
    storeKey      storetypes.StoreKey,
    bankKeeper    bankKeeper.Keeper,
    stakingKeeper stakingKeeper.Keeper,
    accountKeeper accountKeeper.AccountKeeper,
    paramSpace    paramstypes.Subspace,
    authority     string,
) Keeper {
    return Keeper{
        cdc:           cdc,
        storeKey:      storeKey,
        bankKeeper:    bankKeeper,
        stakingKeeper: stakingKeeper,
        accountKeeper: accountKeeper,
        paramSpace:    paramSpace,
        authority:     authority,
    }
}

func (k Keeper) DistributeRewards(sdkCtx sdk.Context, ctx context.Context, rewardAmount sdk.Coins) error {
    proposerAddress := sdkCtx.BlockHeader().ProposerAddress
    if len(proposerAddress) == 0 {
        sdkCtx.Logger().Error("Proposer address is empty")
        return fmt.Errorf("proposer address is empty")
    }

    proposerValidator, _ := k.stakingKeeper.ValidatorByConsAddr(ctx, sdk.ConsAddress(proposerAddress))
    if proposerValidator == nil {
        sdkCtx.Logger().Error("Validator not found", "proposer_address", hex.EncodeToString(proposerAddress))
        return fmt.Errorf("validator not found for proposer address")
    }
    // sdkCtx.Logger().Info("Validator found", "operator_address", proposerValidator.GetOperator())

    // Convert the validator operator address to an account address
    proposerAccAddress, err := sdk.ValAddressFromBech32(proposerValidator.GetOperator())
    if err != nil {
        sdkCtx.Logger().Error("Failed to decode validator operator address", "error", err)
        return fmt.Errorf("failed to decode validator operator address: %w", err)
    }

    accountAddress := sdk.AccAddress(proposerAccAddress)
    sdkCtx.Logger().Info("Proposer Account Address", "account_address", accountAddress.String())

    // Send rewards; if module is drained, this may fail. We log and return gracefully.
    err2 := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, accountAddress, rewardAmount)
    if err2 != nil {
        sdkCtx.Logger().Error("Failed to send block rewards (possibly insufficient funds)", "error", err2, "proposer", accountAddress.String(), "amount", rewardAmount.String())
        return fmt.Errorf("failed to send block rewards: %w", err2)
    }

    sdkCtx.Logger().Info("Distributed block reward", "proposer", accountAddress.String(), "amount", rewardAmount.String())
    return nil
}

// -----------------
// Params (via x/params Subspace)
// -----------------

// GetModuleParams returns the runtime module params from the params subspace, or defaults if unset.
func (k Keeper) GetModuleParams(ctx sdk.Context) types.ModuleParams {
    if k.paramSpace.Name() == "" || !k.paramSpace.Has(ctx, types.KeyRewardAmount) {
        return types.DefaultModuleParams()
    }
    var p types.ModuleParams
    k.paramSpace.GetParamSet(ctx, &p)
    if err := p.Validate(); err != nil {
        return types.DefaultModuleParams()
    }
    return p
}

func (k Keeper) SetModuleParams(ctx sdk.Context, p types.ModuleParams) {
    if k.paramSpace.Name() == "" {
        return
    }
    k.paramSpace.SetParamSet(ctx, &p)
}

// Backwards-compatible Get/Set using protobuf Genesis Params type.
// GetParams returns a protobuf Params with RewardAmount sourced from the param subspace when available.
func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
    if k.paramSpace.Name() != "" && k.paramSpace.Has(ctx, types.KeyRewardAmount) {
        mp := k.GetModuleParams(ctx)
        return types.Params{BlockRewardAmount: mp.RewardAmount}, nil
    }
    // Fallback to legacy KV storage for genesis compatibility if ever used
    store := ctx.KVStore(k.storeKey)
    bz := store.Get([]byte("Params"))
    if bz == nil {
        return types.Params{}, fmt.Errorf("params not found")
    }
    var params types.Params
    if err := k.cdc.Unmarshal(bz, &params); err != nil {
        return types.Params{}, fmt.Errorf("failed to unmarshal Params: %w", err)
    }
    return params, nil
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
    // Prefer writing to param subspace
    if k.paramSpace.Name() != "" {
        p := types.ModuleParams{RewardAmount: params.BlockRewardAmount, HalvingInterval: types.DefaultModuleParams().HalvingInterval}
        k.SetModuleParams(ctx, p)
        return nil
    }
    // Fallback to legacy KV storage
    store := ctx.KVStore(k.storeKey)
    bz, err := k.cdc.Marshal(&params)
    if err != nil {
        return fmt.Errorf("failed to marshal Params: %w", err)
    }
    store.Set([]byte("Params"), bz)
    return nil
}
