package mintburn

import (
    "fmt"

    sdk "github.com/cosmos/cosmos-sdk/types"
    capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
    icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
    channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
    porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
    ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
    ibctmtypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

    mintburn "github.com/maany-xyz/maany-provider/x/mintburn/keeper"
)

// minimal interface for ICA host keeper to resolve ICA address
type ICAHostAddressKeeper interface {
    GetInterchainAccountAddress(ctx sdk.Context, connectionID, portID string) (string, bool)
}

// ICAHostMiddleware wraps the ICA host module to auto-register authorized ICA addresses
// for consumer chains in the mintburn keeper upon channel handshake.
type ICAHostMiddleware struct {
    app       porttypes.IBCModule
    keeper    mintburn.Keeper
    icaHost   ICAHostAddressKeeper
}

func NewICAHostMiddleware(app porttypes.IBCModule, k mintburn.Keeper, host ICAHostAddressKeeper) ICAHostMiddleware {
    return ICAHostMiddleware{app: app, keeper: k, icaHost: host}
}

// OnChanOpenAck intercepts ICA host handshake ack to capture ICA address and map it to the counterparty chain-id
func (im ICAHostMiddleware) OnChanOpenAck(
    ctx sdk.Context,
    portID,
    channelID,
    counterpartyChannelID,
    counterpartyVersion string,
) error {
    if err := im.tryRegisterICA(ctx, portID, channelID); err != nil {
        // log and continue, do not hard-fail handshake
        ctx.Logger().Error("ica_host_mw: register ICA mapping failed", "err", err)
    }
    return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm intercepts confirm as well (depending on who initiated, only one of ack/confirm fires here)
func (im ICAHostMiddleware) OnChanOpenConfirm(
    ctx sdk.Context,
    portID,
    channelID string,
) error {
    if err := im.tryRegisterICA(ctx, portID, channelID); err != nil {
        ctx.Logger().Error("ica_host_mw: register ICA mapping failed", "err", err)
    }
    return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// pass-through for other callbacks
func (im ICAHostMiddleware) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, hops []string, portID, channelID string, cap *capabilitytypes.Capability, cp channeltypes.Counterparty, version string) (string, error) {
    return im.app.OnChanOpenInit(ctx, order, hops, portID, channelID, cap, cp, version)
}

func (im ICAHostMiddleware) OnChanOpenTry(ctx sdk.Context, order channeltypes.Order, hops []string, portID, channelID string, cap *capabilitytypes.Capability, cp channeltypes.Counterparty, cpVersion string) (string, error) {
    return im.app.OnChanOpenTry(ctx, order, hops, portID, channelID, cap, cp, cpVersion)
}

func (im ICAHostMiddleware) OnChanCloseInit(ctx sdk.Context, portID, channelID string) error {
    return im.app.OnChanCloseInit(ctx, portID, channelID)
}

func (im ICAHostMiddleware) OnChanCloseConfirm(ctx sdk.Context, portID, channelID string) error {
    return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

func (im ICAHostMiddleware) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
    return im.app.OnRecvPacket(ctx, packet, relayer)
}

func (im ICAHostMiddleware) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
    return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

func (im ICAHostMiddleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
    return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// internal: resolve controller chain-id and ICA address, then store mapping
func (im ICAHostMiddleware) tryRegisterICA(ctx sdk.Context, portID, channelID string) error {
    // only act for ICA host port
    if portID != icatypes.HostPortID {
        return nil
    }

    ch, found := im.keeper.ChannelKeeper.GetChannel(ctx, portID, channelID)
    if !found {
        return fmt.Errorf("channel not found")
    }
    if len(ch.ConnectionHops) == 0 {
        return fmt.Errorf("missing connection hop")
    }
    connectionID := ch.ConnectionHops[0]
    conn, found := im.keeper.ConnectionKeeper.GetConnection(ctx, connectionID)
    if !found {
        return fmt.Errorf("connection %s not found", connectionID)
    }
    clientID := conn.GetClientID()
    cstate, found := im.keeper.ClientKeeper.GetClientState(ctx, clientID)
    if !found {
        return fmt.Errorf("client state %s not found", clientID)
    }
    tm, ok := cstate.(*ibctmtypes.ClientState)
    if !ok {
        return fmt.Errorf("unexpected client state type")
    }
    consumerChainID := tm.ChainId

    // controller port id is the counterparty port (on controller chain)
    controllerPortID := ch.Counterparty.PortId
    icaAddr, ok := im.icaHost.GetInterchainAccountAddress(ctx, connectionID, controllerPortID)
    if !ok || icaAddr == "" {
        return fmt.Errorf("ica address not found for %s/%s", connectionID, controllerPortID)
    }

    // write mapping
    im.keeper.SetAuthorizedICA(ctx, consumerChainID, icaAddr)
    ctx.Logger().Info("ica_host_mw: registered authorized ICA", "consumer_chain_id", consumerChainID, "ica_addr", icaAddr)
    return nil
}

// no-op
