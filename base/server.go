package base

import (
	randCp "crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/rs/zerolog/log"
	"io"
	"math/rand"
	"net"
	"rilihong/RiServer/proto"
	"strconv"
	"time"
)

const(
	TypeRandom = 0
	TypeHash = 1
	TypeDirect = 2
)

var (
	timeout    = 5 * time.Second
	rTimeout = 10 * time.Second
	ips      = []string{"172.16.75.140:2379", "172.16.75.140:22379", "172.16.75.140:32379"}
	nodeInfo = "{\"ip\":\"127.0.0.1:9880\",\"serverName\":\"routerServer1\",\"serverId\":120001,\"base\":{\"baseName\":\"router\",\"baseId\":120000}}"
)

type HandleFunc func(msg string,session *Session) (string,error)

type Server struct {
	port int
	handleMap map[string]HandleFunc
	gRpcMap map[string]GClientMap		//rpc链接的服务器，发送rpc请求
	randSeed *rand.Rand
	sClientMap map[string]SClientMap	//socket链接的服务器，发送消息
	eServer *EtcdServer
	eCache *EtcdServerCache
}

func (baseServer *Server)registerFunc(str string,handle HandleFunc){
	baseServer.handleMap[str] = handle
}

func (baseServer *Server)BMsgHandle(bMsg []byte,session *Session) ([]byte,error){
	reqStruct :=new(proto.BStruct)
	err := reqStruct.XXX_Unmarshal(bMsg)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("error request struct")
	}
	reqType := reqStruct.Type
	_,ok := baseServer.handleMap[reqType]
	if ok == false{
		log.Error().Str("req",string(bMsg)).Str("type",reqType).Msg("error request type")
	}
	resString,err := baseServer.handleMap[reqType](string(reqStruct.Content),session)
	if err != nil{
		log.Error().Str("req",string(bMsg)).Str("type",reqType).Msg("error request type")
	} else {
		log.Info().Str("req",string(bMsg)).Str("res",resString).Msg("res success")
	}
	return []byte(resString),nil
}

func (baseServer *Server)Init(port int) *Server {
	baseServer.port = port
	baseServer.randSeed = rand.New(rand.NewSource(10))
	baseServer.eServer = NewEtcdServer(ips,timeout)
	err := baseServer.eServer.RegisterServer("/nodes/node1",nodeInfo)
	if err != nil{
		return nil
	}
	sByte := baseServer.eServer.GetAllServer("/nodes/")
	wChan := baseServer.eServer.AddWatch("/nodes/")
	baseServer.eCache.Init(sByte)
	go baseServer.eCache.ListenUpdate(wChan)
	return baseServer
}

func (baseServer *Server) ListenAndServe() {
	l, err := net.Listen("tcp4", ":"+strconv.Itoa(baseServer.port))
	if err != nil {
		//log.Fatal().Err(err)
		panic(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			//log.Fatal().Err(err)
			continue
		}
		sess := NewSession(conn, baseServer)

		go sess.SessionRead()
		go sess.SessionWrite()
	}
}

func (baseServer *Server) SessionId() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(randCp.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (baseServer *Server) GetGClient(bMsg []byte,serverName string,sType int,key int64) (GClient,error){
	var gClient GClient
	var err error
	switch sType {
		case TypeRandom:
			{
				gClient,err = baseServer.gRpcMap[serverName].RandomServer()
				break
			}
		case TypeHash:
			{
				gClient,err = baseServer.gRpcMap[serverName].HashServer(int(key))
				break
			}
		case TypeDirect:
			{
				gClient,err = baseServer.gRpcMap[serverName].DirectServer(key)
				break
			}
		default:
			{
				log.Error().Int("sendType",sType).Msg("sendType error")
				gClient,err = GClient{},errors.New("sendType error")
			}
	}
	if err != nil {
		log.Error().Int("sendType",sType).Str("err",err.Error()).Msg("find client error")
	}
	return gClient,err
}

func (baseServer *Server) GetSClient(bMsg []byte,serverName string,sType int,key int64) (SClient,error){
	var sClient SClient
	var err error
	switch sType {
	case TypeRandom:
		{
			sClient,err = baseServer.sClientMap[serverName].RandomServer()
			break
		}
	case TypeHash:
		{
			sClient,err = baseServer.sClientMap[serverName].HashServer(int(key))
			break
		}
	case TypeDirect:
		{
			sClient,err = baseServer.sClientMap[serverName].DirectServer(key)
			break
		}
	default:
		{
			log.Error().Int("sendType",sType).Msg("sendType error")
			sClient,err = SClient{},errors.New("sendType error")
		}
	}
	if err != nil {
		log.Error().Int("sendType",sType).Str("err",err.Error()).Msg("find client error")
	}
	return sClient,err
}