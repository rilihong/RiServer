package agentserver

import (
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"rilihong/RiServer/base"
	"rilihong/RiServer/proto"
	"rilihong/RiServer/src/agent/userinfo"
)

type AgentServer struct {
	base.Server
	userMap map[string]UserInfo.UserInfo	//session + UserInfo
	IdToSession map[int64]string
}

func NewAgentServer(port int) *AgentServer{
	agentServer := new(AgentServer)
	agentServer.Init(port)
	return agentServer
}

func (server *AgentServer)BMsgHandle(bMsg []byte,session *base.Session) ([]byte,error){
	agentMsg := proto.AgentReq{ReqType: "agent_pass", ReqContent:bMsg,ReqToken:base.RandToken()}
	tmpMsg := make([]byte,1024)
	b,err := agentMsg.XXX_Marshal(tmpMsg,false)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("add json err")
	}else{
		err = server.SendServerMessage(tmpMsg,"RouterServer",base.TypeRandom,0)
	}
	return nil,err
}

// 发送socket消息
func (server *AgentServer)SendServerMessage(bMsg []byte,serverName string,rType int,key int64) error{
	sClient,err := server.GetSClient(bMsg,serverName,rType,key)
	if err != nil {
		err = server.TryConnect("RouterServer")
		if err != nil {
			return errors.New("no server")
		}
		sClient,err = server.GetSClient(bMsg,serverName,rType,key)
		if err != nil {
			return errors.New("no client")
		}
	}
	sClient.Session.Write(bMsg)
	return nil
}

// 重新链接服务
func (server *AgentServer)TryConnect(serverName string) error{
	return nil
}