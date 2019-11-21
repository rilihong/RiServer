package base

import (
	"context"
	"fmt"
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
	lease := clientv3.NewLease(eServer.client)
	leaseRes,err := lease.Grant(context.TODO(),2)
	if err != nil{
		return err
	}
	leaseId := leaseRes.ID
	//自动续租（底层会每次讲租约信息扔到 <-chan *clientv3.LeaseKeepAliveResponse 这个管道中）
	keepRespChan,err := lease.KeepAlive(context.TODO(),leaseId)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//启动一个新的协程来select这个管道
	go func() {
		for {
			select {
			case <- keepRespChan:
			}
		}
	}()

	//kv := clientv3.NewKV(eServer.client)
	//进行写操作
	//if _,err := kv.Put(context.TODO(),key,serverInfo,clientv3.WithLease(leaseId));err != nil {
	//	fmt.Println(err)
	//return err
	//}
	_,err = eServer.client.Put(context.TODO(), key,serverInfo,clientv3.WithLease(leaseId))
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