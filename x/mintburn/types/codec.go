package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func RegisterInterfaces(reg codectypes.InterfaceRegistry) {
    reg.RegisterImplementations((*sdk.Msg)(nil),
        &MsgEscrowInitial{},
        &MsgCancelEscrow{},
        &MsgMarkEscrowClaimed{},
    )
}
