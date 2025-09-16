package mintburn

import (
    "encoding/json"
    "fmt"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"

	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
    mintburn "github.com/maany-xyz/maany-provider/x/mintburn/keeper"
    mintburntypes "github.com/maany-xyz/maany-provider/x/mintburn/types"
)

type IBCMiddleware struct {
	app    porttypes.IBCModule
	keeper mintburn.Keeper
}

func NewIBCMiddleware(app porttypes.IBCModule, k mintburn.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:    app,
		keeper: k,
	}
}

// OnChanOpenInit implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, version)
}

// OnChanOpenTry implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	// call underlying app's OnChanOpenTry callback with the appVersion
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, counterpartyVersion)
}

func (im IBCMiddleware) HandleChannelIdStorage(
	ctx sdk.Context,
	portID,
	channelID string,
	isOpening bool,
) error {
	if portID != "transfer" {
		return nil
	}

	channel, found := im.keeper.ChannelKeeper.GetChannel(ctx, portID, channelID)
	if !found {
		return fmt.Errorf("channel not found")
	}
	connectionID := channel.ConnectionHops[0]
	connection, found := im.keeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return fmt.Errorf("connection %s not found", connectionID)
	}
	clientID := connection.Counterparty.ClientId
	clientState, found := im.keeper.ClientKeeper.GetClientState(ctx, clientID)
	if !found {
		return fmt.Errorf("client state for %s not found", clientID)
	}
	tmClientState, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		return fmt.Errorf("unexpected client state type")
	}
    // Chain-id allowlist: only allow channels for explicitly allowed chain-ids.
    // For now, also allow hardcoded "maanydex" until params/genesis wiring exists.
    counterpartyChainID := tmClientState.ChainId
    allowedChain := (counterpartyChainID == "maanydex") || im.keeper.IsAllowedConsumerChain(ctx, counterpartyChainID)
    if !allowedChain {
        ctx.Logger().Info("mintburn: ignoring transfer channel; counterparty chain-id not allow-listed",
            "channel_id", channelID, "counterparty_chain_id", counterpartyChainID)
        return nil
    }
    if isOpening {
        store := prefix.NewStore(ctx.KVStore(im.keeper.StoreKey), []byte("allowed-channel/"))
        store.Set([]byte(channelID), []byte{1})
        // also track mapping chain-id -> channel for later escrow funding
        im.keeper.AddAllowedChannelForChain(ctx, counterpartyChainID, channelID)
        ctx.Logger().Info("mintburn: allow-listed transfer channel",
            "channel_id", channelID, "counterparty_chain_id", counterpartyChainID)
    } else {
        store := prefix.NewStore(ctx.KVStore(im.keeper.StoreKey), []byte("allowed-channel/"))
        store.Delete([]byte(channelID))
        im.keeper.RemoveAllowedChannelForChain(ctx, counterpartyChainID, channelID)
        ctx.Logger().Info("mintburn: removed transfer channel from allow-list",
            "channel_id", channelID, "counterparty_chain_id", counterpartyChainID)
    }

	return nil
}

func (im IBCMiddleware) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID,
	counterpartyChannelID,
	counterpartyVersion string,
) error {
	// Run middleware logic first
	if err := im.HandleChannelIdStorage(ctx, portID, channelID, true); err != nil {
		//On any error concerning establishing a transfer channel, return the error and aboart the channel creation 
		return err
	}
	// Then forward to the underlying IBC app
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}
 
// OnChanOpenConfirm implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	//Note: since a channel can be opened from both ways, need to handle same logic here as for OnChanOpenAck
	// Run middleware logic first
	if err := im.HandleChannelIdStorage(ctx, portID, channelID, true); err != nil {
		//On any error concerning establishing a transfer channel, return the error and aboart the channel creation 
		return err
	}
	// Then forward to the underlying IBC app
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCMiddleware interface
func (im IBCMiddleware) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Run middleware logic first
	if err := im.HandleChannelIdStorage(ctx, portID, channelID, false); err != nil {
	//On any error concerning establishing a transfer channel, return the error and aboart the channel creation 
		return err
	}
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

func (im IBCMiddleware) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Run middleware logic first
	if err := im.HandleChannelIdStorage(ctx, portID, channelID, false); err != nil {
	//On any error concerning establishing a transfer channel, return the error and aboart the channel creation 
		return err
	}
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnTimeoutPacket implements the IBCMiddleware interface
// If fees are not enabled, this callback will default to the ibc-core packet callback
func (im IBCMiddleware) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	// call underlying app's OnTimeoutPacket callback.
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// Provider middleware: OnRecvPacket (for DEX->Provider)
func (im IBCMiddleware) OnRecvPacket(
    ctx sdk.Context,
    packet channeltypes.Packet,
    relayer sdk.AccAddress,
) exported.Acknowledgement {
    okAck := channeltypes.NewResultAcknowledgement([]byte{byte(1)})

    var data ibctransfertypes.FungibleTokenPacketData
    if err := json.Unmarshal(packet.GetData(), &data); err != nil {
        ctx.Logger().Error("mintburn: invalid packet data", "err", err)
        return channeltypes.NewErrorAcknowledgement(fmt.Errorf("invalid packet data"))
    }

    // Log any incoming transfer packet on recv path (consumer -> provider)
    ctx.Logger().Info("mintburn: recv transfer packet",
        "src_port", packet.SourcePort, "src_channel", packet.SourceChannel,
        "dst_port", packet.DestinationPort, "dst_channel", packet.DestinationChannel,
        "sender", data.Sender, "receiver", data.Receiver,
        "denom", data.Denom, "amount", data.Amount,
    )

    // Only handle our allow-listed transfer channel (use local identifiers on recv)
    if packet.DestinationPort != "transfer" || !im.keeper.IsAllowedChannel(ctx, packet.DestinationChannel) {
        ctx.Logger().Info("mintburn: skip recv — not allow-listed or not transfer port",
            "dst_port", packet.DestinationPort, "dst_channel", packet.DestinationChannel)
        return im.app.OnRecvPacket(ctx, packet, relayer)
    }

    // Resolve base denom and verify it matches your provider base (e.g., "umaany").
    // ReceiverChainIsSource must use local (destination) identifiers on recv.
    baseDenom := data.Denom
    if !ibctransfertypes.ReceiverChainIsSource(packet.DestinationPort, packet.DestinationChannel, data.Denom) {
        trace := ibctransfertypes.ParseDenomTrace(data.Denom)
        baseDenom = trace.BaseDenom
    }
    ctx.Logger().Info("mintburn: resolved base denom", "base_denom", baseDenom)
    if baseDenom != "umaany" { // or "stake" depending on your mainnet minimal denom
        // Not our path -> let transfer app mint a voucher
        ctx.Logger().Info("mintburn: skip recv — base denom not provider base", "base_denom", baseDenom)
        return im.app.OnRecvPacket(ctx, packet, relayer)
    }

    // Validate receiver and amount
    rcpt, err := sdk.AccAddressFromBech32(data.Receiver)
    if err != nil {
        ctx.Logger().Error("mintburn: invalid receiver", "receiver", data.Receiver, "err", err)
        return channeltypes.NewErrorAcknowledgement(fmt.Errorf("invalid receiver"))
    }
    amt, ok := sdkmath.NewIntFromString(data.Amount)
    if !ok || !amt.IsPositive() {
        ctx.Logger().Error("mintburn: invalid amount", "amount", data.Amount)
        return channeltypes.NewErrorAcknowledgement(fmt.Errorf("invalid amount"))
    }

    // RELEASE from the Provider escrow to the recipient (no voucher mint!) — use the dedicated mintburn escrow address
    // Derive the consumer chain-id from the local channel/connection/client.
    ch, found := im.keeper.ChannelKeeper.GetChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
    if !found || len(ch.ConnectionHops) == 0 {
        ctx.Logger().Error("mintburn: recv on unknown channel", "dst_port", packet.DestinationPort, "dst_channel", packet.DestinationChannel)
        return channeltypes.NewErrorAcknowledgement(fmt.Errorf("unknown channel"))
    }
    conn, found := im.keeper.ConnectionKeeper.GetConnection(ctx, ch.ConnectionHops[0])
    if !found {
        return channeltypes.NewErrorAcknowledgement(fmt.Errorf("unknown connection"))
    }
    cstate, found := im.keeper.ClientKeeper.GetClientState(ctx, conn.Counterparty.ClientId)
    if !found {
        return channeltypes.NewErrorAcknowledgement(fmt.Errorf("unknown client state"))
    }
    tm, ok := cstate.(*ibctmtypes.ClientState)
    if !ok {
        return channeltypes.NewErrorAcknowledgement(fmt.Errorf("unexpected client state type"))
    }

    mintburnEscrowID := mintburntypes.MintburnEscrowChannelIDFor(tm.ChainId)
    escrowAddr := ibctransfertypes.GetEscrowAddress("transfer", mintburnEscrowID)
    coin := sdk.NewCoin("umaany", amt)

    ctx.Logger().Info("mintburn: releasing from escrow",
        "amount", coin.String(), "to", rcpt.String(), "escrow", escrowAddr.String())

    if err := im.keeper.SendFromEscrowToAccount(ctx, escrowAddr, rcpt, coin); err != nil {
        ctx.Logger().Error("mintburn: release from escrow failed", "err", err)
        return channeltypes.NewErrorAcknowledgement(fmt.Errorf("release failed"))
    }

    ctx.Logger().Info("mintburn: released from provider escrow to recipient",
        "amount", coin.String(), "to", rcpt.String(), "escrow", escrowAddr.String())

    // IMPORTANT: do NOT forward to transfer app; return success ack
    return okAck
}


func (im IBCMiddleware) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	ctx.Logger().Info("mintburn/provider: ack passthrough",
		"src_port", packet.SourcePort, "src_channel", packet.SourceChannel,
		"dst_port", packet.DestinationPort, "dst_channel", packet.DestinationChannel,
		"seq", packet.Sequence)
	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)

}
