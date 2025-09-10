package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
}

// NewKeeper creates a new blockrewards Keeper instance
func NewKeeper (
    cdc           codec.BinaryCodec,
    storeKey      storetypes.StoreKey,
    bankKeeper    bankKeeper.Keeper,
    stakingKeeper stakingKeeper.Keeper,
    accountKeeper accountKeeper.AccountKeeper,
) Keeper {
    return Keeper{
        cdc:           cdc,
        storeKey:      storeKey,
        bankKeeper:    bankKeeper,
        stakingKeeper: stakingKeeper,
        accountKeeper: accountKeeper,
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

    // Send rewards
    err2 := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, "blockrewards", accountAddress, rewardAmount)
    if err2 != nil {
        sdkCtx.Logger().Error("Failed to send block rewards", "error", err2, "proposer", accountAddress.String(), "amount", rewardAmount.String())
        return fmt.Errorf("failed to send block rewards: %w", err2)
    }

    sdkCtx.Logger().Info("Distributed block reward", "proposer", accountAddress.String(), "amount", rewardAmount.String())
    return nil
}

func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
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
    store := ctx.KVStore(k.storeKey)
    bz, err := k.cdc.Marshal(&params)
    if err != nil {
        return fmt.Errorf("failed to marshal Params: %w", err)
    }

    store.Set([]byte("Params"), bz)
    return nil
}