package main

import (
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"net"
	"rilihong/RiServer/proto"
	"rilihong/RiServer/src/chinese_poker/pokerserver"
)

const (
	port = ":50051"
)

func main(){
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	s := grpc.NewServer()
	p := pokerserver.NewPokerServer(19408)
	proto.RegisterPokerServerServer(s,p)
	if err := s.Serve(lis); err != nil {
		log.Fatal().Err(err).Send()
	}


}