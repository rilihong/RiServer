package redis

import (
	"github.com/go-redis/redis"
	"time"
)

type Client struct{
	_op Opt
	_client *redis.Client
	_redisEnd chan struct{}
}

type Opt struct{
	IP 			string
	Port 		int
	Auth 		string
	PoolSize  	int
	DB			int
	TimeOut 	time.Duration
}

func (op *Opt)Init(){
	if op.IP == "" {
		op.IP = "127.0.0.1"
	}
	if op.Port == 0 {
		op.Port = 6379
	}
	if op.PoolSize == 0 {
		op.PoolSize = 10
	}
	if op.DB < 0 || op.DB > 16 {
		op.DB = 0
	}
	if op.TimeOut <= 0 {
		op.TimeOut = 10 * time.Second
	}
}

func CopyOpt(op Opt) *redis.Options {
	rOpt := new(redis.Options)
	rOpt.Addr = op.IP
	rOpt.Password = op.Auth
	rOpt.DB = op.DB
	rOpt.PoolSize = op.PoolSize
	rOpt.DialTimeout = op.TimeOut
	return rOpt
}

func NewRedis(op Opt) *Client{
	cli := redis.NewClient(CopyOpt(op))
	if cli == nil{
		return nil
	}
	client := new(Client)
	client._client = cli
	client._op = op
	client._redisEnd = make(chan struct{})
	go KeepALive(client)
	return client
}

func (client *Client)Get(key string) (string ,error){
	return client._client.Get(key).Result()
}

func (client *Client)Set(key string, value interface{},
								expiration time.Duration) (bool,error){
	_,err :=client._client.Set(key,value,expiration).Result()
	if err != nil {
		return false,err
	}
	return true,nil
}

func KeepALive(client *Client){
	var f func()
	f = func(){
		select{
			case <- client._redisEnd :
				return
			default:
		}
		_,err := client._client.Ping().Result()
		if err != nil{
			return
		}
		time.AfterFunc(time.Second * 30,f)
	}
	f()
}

func (client *Client)Close(){
	client._client.Close()
	client._redisEnd <- struct{}{}
}