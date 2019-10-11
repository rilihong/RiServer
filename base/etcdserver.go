package base

import (
	"context"
	"github.com/rs/zerolog/log"
	"go.etcd.io/etcd/clientv3"
	"time"
)

type EtcdServer struct{
	client *clientv3.Client
}

func NewEtcdServer(ips []string,timeout time.Duration) *EtcdServer{
	server := new(EtcdServer)
	server.Init(ips ,timeout)
	return server
}

func (eServer *EtcdServer)Init(ips []string,timeout time.Duration){
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   ips,
		DialTimeout: timeout,
	})
	if err != nil {
		log.Error().Msg(err.Error())
	}
	eServer.client = cli
}

func (eServer *EtcdServer)AddWatch(nodePrefix string) clientv3.WatchChan{
	return eServer.client.Watch(context.TODO(), "/nodes",clientv3.WithPrefix())
}

func (eServer *EtcdServer)RegisterServer(key string,serverInfo string) error{
	_,err := eServer.client.Put(context.TODO(), key,serverInfo,nil)
	return err
}

func (eServer *EtcdServer)GetAllServer(key string) []SByte{
	gRes,err := eServer.client.Get(context.TODO(), key,clientv3.WithPrefix())
	if err != nil{
		return []SByte{}
	}

	var sByteList = make([]SByte,len(gRes.Kvs))
	for i,v := range gRes.Kvs{
		sByteList[i] = v.Value
	}
	return sByteList
}