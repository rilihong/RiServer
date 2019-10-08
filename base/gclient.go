package base

import "github.com/grpc/grpc-go"

type GClient struct{
	cc *grpc.ClientConn
}
