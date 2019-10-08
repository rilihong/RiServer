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
)

type HandleFunc func(msg string,session *Session) (string,error)

type GClientMap map[int]GClient

type Server struct {
	port int
	handleMap map[string]HandleFunc
	gRpcMap map[string]GClientMap
	rand_seed *rand.Rand
}

func (baseServer *Server)registerFunc(str string,handle HandleFunc){
	baseServer.handleMap[str] = handle
}

func (baseServer *Server)BMsgHandle(bMsg []byte,session *Session) ([]byte,error){
	reqStruct :=new(Struct)
	err := json.Unmarshal(bMsg,*reqStruct)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("error request struct")
	}
	reqType := reqStruct.ReqType
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
	baseServer.rand_seed = rand.New(rand.NewSource(10))
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

func (baseServer *Server) RandSend(bMsg []byte,serverName string) error{
	idMap,ok := baseServer.gRpcMap[serverName]
	if ok == false{
		return errors.New("no_server")
	}
	idLen := len(idMap)
	if idLen == 0{
		delete(idMap, serverName)
		return errors.New("server_len_zero")
	}
	index := baseServer.rand_seed.Int()%idLen

	return nil
}