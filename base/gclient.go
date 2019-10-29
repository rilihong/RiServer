package base

import (
	"errors"
	"google.golang.org/grpc"
	"math/rand"
)

type GClient struct{
	cc *grpc.ClientConn
}

type GClientMap struct {
	gMap map[int64]GClient
	baseNum int64
}

func (gCMap GClientMap)HashServer(key int) (GClient,error){
	gLen := len(gCMap.gMap)
	if gLen == 0 {
		return GClient{},errors.New("ServerLen_zero")
	}
	return gCMap.gMap[gCMap.baseNum + int64(key%gLen)],nil
}

func (gCMap GClientMap)DirectServer(serverId int64)  (GClient,error){
	gClient,ok := gCMap.gMap[serverId]
	if ok == false{
		return GClient{},errors.New("ServerId_not_exist")
	}
	return gClient,nil
}

func (gCMap GClientMap)RandomServer()  (GClient,error){
	gLen := len(gCMap.gMap)
	if gLen == 0 {
		return GClient{},errors.New("ServerLen_zero")
	}
	index := rand.Int()%gLen
	count := 0
	for _,v := range gCMap.gMap {
		if count == index{
			return v,nil
		} else{
			count++
		}
	}
	return GClient{},errors.New("server_not_exist")
}