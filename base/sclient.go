package base

import (
	"errors"
	"github.com/rs/zerolog/log"
	"math/rand"
	"net"
)

type SClient struct{
	Session *Session
	ServerId int64
	ServerName string
}

func NewSClient(ip string,serverId int64,serverName string,server *Server ) *SClient{
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		log.Error().Str("ip",ip).Msg("connect error")
	}
	sClient := new(SClient)
	sClient.Session = NewSession(conn,server,server.routerHandle)
	sClient.ServerId = serverId
	sClient.ServerName = serverName

	go sClient.Session.SessionRead()
	go sClient.Session.SessionWrite()
	return sClient
}

type SClientMap struct {
	gMap map[int64]SClient
	baseNum int64
}

func (sCMap SClientMap)HashServer(key int) (SClient,error){
	gLen := len(sCMap.gMap)
	if gLen == 0 {
		return SClient{},errors.New("ServerLen_zero")
	}
	return sCMap.gMap[sCMap.baseNum + int64(key%gLen)],nil
}

func (sCMap SClientMap)DirectServer(serverId int64)  (SClient,error){
	gClient,ok := sCMap.gMap[serverId]
	if ok == false{
		return SClient{},errors.New("ServerId_not_exist")
	}
	return gClient,nil
}

func (sCMap SClientMap)RandomServer()  (SClient,error){
	sLen := len(sCMap.gMap)
	if sLen == 0 {
		return SClient{},errors.New("ServerLen_zero")
	}
	index := rand.Int()%sLen
	count := 0
	for _,v := range sCMap.gMap {
		if count == index{
			return v,nil
		} else{
			count++
		}
	}
	return SClient{},errors.New("server_not_exist")
}