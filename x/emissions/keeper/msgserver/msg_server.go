package msgserver

import (
	"github.com/allora-network/allora-chain/x/emissions/types"
	"github.com/allora-network/allora-chain/x/emissions/keeper"
)

type msgServer struct {
	k keeper.Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the module MsgServer interface.
func NewMsgServerImpl(keeper keeper.Keeper) types.MsgServer {
	return &msgServer{k: keeper}
}
