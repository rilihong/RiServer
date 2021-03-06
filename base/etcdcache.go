package base

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"go.etcd.io/etcd/clientv3"
	"sync"
	"time"
)

var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 10 * time.Second
	endpoints      = []string{"120.92.42.105:2379","120.92.42.105:32379","120.92.86.87:22379"}
)

type ServerBase struct{
	BaseName string 	`json:"baseName"`
	BaseId int64 		`json:"baseId"`
}

func (bInfo ServerBase)String() string{
	return fmt.Sprintf("{baseName %s,baseId %d}",bInfo.BaseName,bInfo.BaseId)
}

type ServerInfo struct {
	ServerName string 	`json:"serverName"`
	ServerId int64 		`json:"serverId"`
	Ip string 			`json:"ip"`
	SBase ServerBase	`json:"base"`
}

func (sInfo ServerInfo)String() string{
	return fmt.Sprintf("{ServerName %s,ServerId %d Ip %s base %s}",sInfo.ServerName,sInfo.ServerId,sInfo.Ip,sInfo.SBase)
}

type SameServerMap struct {
	BaseName string
	BaseId int64
	ServerIdMap map[int64] ServerInfo
}

type EtcdServerCache struct {
	ServerMap map[string]SameServerMap		//serverName to serverInfo
	IdMap	map[int64]ServerInfo			//serverID to serverInfo
	RWLock *sync.RWMutex						// read write lock
}

func NewEtcdServerCache(allServer []SByte) *EtcdServerCache{
	eCache := new(EtcdServerCache)
	eCache.ServerMap = make(map[string]SameServerMap)		//serverName to serverInfo
	eCache.IdMap = make(map[int64]ServerInfo)			//serverID to serverInfo
	eCache.RWLock = new(sync.RWMutex)
	eCache.Init(allServer)
	return eCache
}

func (eSCache *EtcdServerCache)Init(allServer []SByte){
	for _,v := range allServer{
		eSCache.AddServerInfo(v)
		fmt.Println(" init :",string(v))
	}
}

func (eSCache *EtcdServerCache)ListenUpdate(wChan clientv3.WatchChan){
	for {
		select {
		case ev := <-wChan:
			for _, v := range ev.Events {
				fmt.Println("type :", v.Type, " k:", string(v.Kv.Key), " v:", string(v.Kv.Value))
				eSCache.AddServerInfo(v.Kv.Value)
			}
		}
	}
}

func (eSCache *EtcdServerCache)AddServerInfo(sByte []byte) error{
	var sInfo ServerInfo
	err := json.Unmarshal(sByte, &sInfo)
	if err != nil {
		log.Error().Str("error", err.Error()).Send()
		return err
	}
	{
		eSCache.RWLock.Lock()
		s,ok := eSCache.ServerMap[sInfo.SBase.BaseName]
		if ok != true{
			var sSM = SameServerMap{BaseId:sInfo.SBase.BaseId,BaseName:sInfo.SBase.BaseName,ServerIdMap:make(map[int64] ServerInfo)}
			eSCache.ServerMap[sInfo.SBase.BaseName] = sSM
			s = eSCache.ServerMap[sInfo.SBase.BaseName]
		}
		s.ServerIdMap[sInfo.ServerId] = sInfo
		s.BaseName = sInfo.SBase.BaseName
		s.BaseId = sInfo.SBase.BaseId
		eSCache.IdMap[sInfo.ServerId] = sInfo
		eSCache.RWLock.Unlock()
	}
	return nil
}

func (eSCache *EtcdServerCache)GetServerInfoById(serverId int64) (*ServerInfo,error){
	eSCache.RWLock.Lock()
	defer eSCache.RWLock.Unlock()
	s,ok := eSCache.IdMap[serverId]
	if ok != true{
		return nil,errors.New("no server info")
	}
	return &s,nil
}

func (eSCache *EtcdServerCache)GetServerInfoByName(serverName string) SameServerMap{
	eSCache.RWLock.Lock()
	defer eSCache.RWLock.Unlock()
	s,ok := eSCache.ServerMap[serverName]
	if ok != true{
		fmt.Println("eSCache no info")
		return SameServerMap{}
	}
	fmt.Println("eSCache has info")
	return s
}

