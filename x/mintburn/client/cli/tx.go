package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	mintburntypes "github.com/maany-xyz/maany-provider/x/mintburn/types"
)

func NewTxCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use:                        "mintburn",
    Short:                      "Mintburn transactions",
    DisableFlagParsing:         true,
    SuggestionsMinimumDistance: 2,
    RunE:                       client.ValidateCmd,
  }
  cmd.AddCommand(
    NewEscrowInitialCmd(),
    NewCancelEscrowCmd(),
    NewMarkEscrowClaimedCmd(),
  )
  return cmd
}

func NewEscrowInitialCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use:   "escrow-initial [consumer-chain-id] [amount] [recipient(optional)]",
    Short: "Lock provider base tokens into escrow",
    Args:  cobra.RangeArgs(2, 3),
    RunE: func(cmd *cobra.Command, args []string) error {
      clientCtx, err := client.GetClientTxContext(cmd)
      if err != nil { return err }

      consumerID := args[0]
      coin, err := sdk.ParseCoinNormalized(args[1]) // e.g. 1000000stake
      if err != nil { return err }

      recipient := ""
      if len(args) == 3 {
        recipient = args[2]
      }

      expiryH, _ := cmd.Flags().GetUint64("expiry-height")
      expiryT, _ := cmd.Flags().GetUint64("expiry-time-unix")

      msg := &mintburntypes.MsgEscrowInitial{
        Sender:          clientCtx.GetFromAddress().String(),
        ConsumerChainId: consumerID,
        Amount:          coin,
        Recipient:       recipient,
        ExpiryHeight:    expiryH,
        ExpiryTimeUnix:  expiryT,
      }
      if err := msg.ValidateBasic(); err != nil { return err }
      return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
    },
  }
  flags.AddTxFlagsToCmd(cmd)
  cmd.Flags().Uint64("expiry-height", 0, "Optional expiry height")
  cmd.Flags().Uint64("expiry-time-unix", 0, "Optional expiry UNIX time")
  return cmd
}

func NewMarkEscrowClaimedCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use:   "mark-escrow-claimed [escrow-id] [consumer-chain-id]",
    Short: "Mark an escrow as claimed by its id",
    Args:  cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
      clientCtx, err := client.GetClientTxContext(cmd)
      if err != nil { return err }

      msg := &mintburntypes.MsgMarkEscrowClaimed{
        Sender:   clientCtx.GetFromAddress().String(),
        EscrowId: args[0],
        ConsumerChainId: args[1],
      }
      if err := msg.ValidateBasic(); err != nil { return err }
      return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
    },
  }
  flags.AddTxFlagsToCmd(cmd)
  return cmd
}

func NewCancelEscrowCmd() *cobra.Command {
  cmd := &cobra.Command{
    Use:   "cancel-escrow [consumer-chain-id] [denom]",
    Short: "Cancel a pending escrow and refund",
    Args:  cobra.ExactArgs(2),
    RunE: func(cmd *cobra.Command, args []string) error {
      clientCtx, err := client.GetClientTxContext(cmd)
      if err != nil { return err }

      msg := &mintburntypes.MsgCancelEscrow{
        Sender:          clientCtx.GetFromAddress().String(),
        ConsumerChainId: args[0],
        Denom:           args[1],
      }
      if err := msg.ValidateBasic(); err != nil { return err }
      return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
    },
  }
  flags.AddTxFlagsToCmd(cmd)
  return cmd
}
