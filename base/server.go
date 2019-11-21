package base

import (
	randCp "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"io"
	"math/rand"
	"net"
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
	ips      = []string{"120.92.42.105:2379","120.92.42.105:32379","120.92.86.87:22379"}
	nodeInfo = "{\"ip\":\"127.0.0.1:9880\",\"serverName\":\"routerServer1\",\"serverId\":120001,\"base\":{\"baseName\":\"router\",\"baseId\":120000}}"
)

type MsgHandle func(bMsg []byte,session *Session) ([]byte,error)

type HandleFunc func(bMsg []byte,session *Session) ([]byte,string,error)

type Server struct {
	port int
	baseName string
	serverId int64
	serverName string
	handleMap map[string]HandleFunc
	gRpcMap map[string]GClientMap		//rpc链接的服务器，发送rpc请求
	randSeed *rand.Rand
	sClientMap map[string]SClientMap	//socket链接的服务器，发送消息
	eServer *EtcdServer
	ECache *EtcdServerCache
	handle MsgHandle
	routerHandle MsgHandle
}

func (baseServer *Server)RegisterFunc(str string,handle HandleFunc){
	baseServer.handleMap[str] = handle
}

func (baseServer *Server)GetFunc(str string) (HandleFunc,error){
	handle,ok := baseServer.handleMap[str]
	if ok != true{
		return nil,errors.New("func not exist")
	}
	return handle,nil
}

func (baseServer *Server)SetMsgHandle(h MsgHandle){
	baseServer.handle = h
}

func (baseServer *Server)SetRouterHandle(h MsgHandle){
	baseServer.routerHandle = h
}

func (baseServer *Server)BMsgHandle(bMsg []byte,session *Session) ([]byte,error){
	return nil,nil
}

func (baseServer *Server)Init(port int,baseName string,serverId int64) *Server {
	baseServer.port = port
	baseServer.serverId = serverId
	baseServer.baseName = baseName
	baseServer.serverName = baseName + strconv.FormatInt(baseServer.serverId,10)
	baseServer.handleMap = make(map[string]HandleFunc)
	baseServer.gRpcMap = make(map[string]GClientMap)	//rpc链接的服务器，发送rpc请求
	baseServer.sClientMap = make(map[string]SClientMap)	//socket链接的服务器，发送消息
	baseServer.randSeed = rand.New(rand.NewSource(10))
	baseServer.eServer = NewEtcdServer(ips,timeout)
	baseServer.SetMsgHandle(baseServer.BMsgHandle)
	var sInfo = ServerInfo{
				ServerName:baseServer.serverName,
				ServerId:baseServer.serverId,
				Ip:"127.0.0.1" + ":" + strconv.Itoa(baseServer.port),
				SBase:ServerBase{
					BaseName:baseServer.baseName,
					BaseId:baseServer.serverId - (baseServer.serverId%1000)}}
	etcdByte,_ := json.Marshal(sInfo)
	err := baseServer.eServer.RegisterServer("/nodes/" + baseServer.baseName + "/"+ strconv.FormatInt(baseServer.serverId,10) ,string(etcdByte))
	if err != nil{
		return nil
	}
	sByte := baseServer.eServer.GetAllServer("/nodes/")
	wChan := baseServer.eServer.AddWatch("/nodes/")
	baseServer.ECache = NewEtcdServerCache(sByte)
	go baseServer.ECache.ListenUpdate(wChan)
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
		sess := NewSession(conn,baseServer.handle)

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

// 重新链接服务
func (baseServer *Server)TryConnect(serverName string) error{
	sMap := baseServer.ECache.GetServerInfoByName(serverName)
	for _,v := range sMap.ServerIdMap{
		cli := NewSClient(v.Ip,v.ServerId,v.ServerName,baseServer.routerHandle)
		if cli != nil {
			value,ok := baseServer.sClientMap[v.SBase.BaseName]
			if ok != true{
				var ss SClientMap
				ss.gMap = make(map[int64]SClient)
				baseServer.sClientMap[v.SBase.BaseName] = ss
				value = baseServer.sClientMap[v.SBase.BaseName]
			}
			value.baseNum = v.ServerId - v.ServerId%1000
			value.gMap[v.ServerId] = *cli
		}
	}
	return nil
}