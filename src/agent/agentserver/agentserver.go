package agentserver

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"rilihong/RiServer/base"
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
	agentMsg := base.Struct{ReqType: "agent_pass", Content:bMsg,SessionId:session.SessionId()}
	b,err := json.Marshal(agentMsg)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("add json err")
	}
	return server.SendRouter(b)
}

func (server *AgentServer)SendRouter(bMsg []byte) ([]byte,error){
	return nil,nil
}