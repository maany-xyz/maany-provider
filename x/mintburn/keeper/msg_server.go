package mintburn

type MsgServer struct {
    Keeper
}

func NewMsgServer(keeper Keeper) MsgServer {
    return MsgServer{Keeper: keeper}
}
