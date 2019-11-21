package main

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"rilihong/RiServer/proto"
	"time"
)

const (
	address     = "127.0.0.1:50051"
)

func main(){
	conn, err := grpc.Dial(address,grpc.WithInsecure())
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer conn.Close()
	c := proto.NewPokerServerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetPoker(ctx,&proto.PokerReq{Uid:1024,Name:"kitty"})
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	log.Printf("Greeting: %s", r.String())
}
