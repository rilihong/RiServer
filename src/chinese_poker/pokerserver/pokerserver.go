package pokerserver

import (
	"context"
	"github.com/rs/zerolog/log"
	"rilihong/RiServer/base"
	"rilihong/RiServer/proto"
)

type PokerServer struct {
	base.Server
}

func NewPokerServer(port int) *PokerServer{
	poker := new(PokerServer)
	poker.Init(port,"poker",1940001)
	//poker.SetMsgHandle(routerServer.BMsgHandle)
	//poker.RegisterFunc("login",routerServer.AgentPass)
	return poker
}

func (s *PokerServer) GetPoker(ctx context.Context, in *proto.PokerReq) (*proto.PokerRes, error) {
	log.Info().Str("name",in.Name)
	res := &proto.PokerRes{Uid:in.Uid,Name:in.Name,Result:"ok",Table:1}
	return res, nil
}