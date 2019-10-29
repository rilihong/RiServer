package agentserver

import (
	"errors"
	"fmt"
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
	agentServer.Init(port,"agent",1920001)
	agentServer.SetMsgHandle(agentServer.BMsgHandle)
	agentServer.SetRouterHandle(agentServer.RouterMsgHandle)
	agentServer.RegisterFunc("login",agentServer.NormalHandle)
	agentServer.RegisterFunc("agent_receive",agentServer.RouterResHandle)
	return agentServer
}

func (server *AgentServer)NormalHandle(bMsg []byte,session *base.Session) ([]byte,error){
	reqStruct :=new(proto.BStruct)
	err := reqStruct.XXX_Unmarshal(bMsg)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("NormalHandle error request struct")
		return nil,err
	}
	fmt.Println("type:",reqStruct.GetType()," content:",string(reqStruct.GetContent()))
	bLen := len(reqStruct.GetContent())
	res := make([]byte,bLen)
	copy(res,reqStruct.GetContent())
	return res,nil
}

func (server *AgentServer)RouterResHandle(bMsg []byte,session *base.Session) ([]byte,error){

	agentRes :=new(proto.AgentRes)
	err := agentRes.XXX_Unmarshal(bMsg)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("RouterHandle error request struct")
		return nil,err
	}
	fmt.Println("type:",agentRes.GetResType()," content:",string(agentRes.GetResContent()))
	bLen := len(agentRes.GetResContent())
	res := make([]byte,bLen)
	copy(res,agentRes.GetResContent())
	return res,nil
}

func (server *AgentServer)BMsgHandle(bMsg []byte,session *base.Session) ([]byte,error){
	resString := &proto.BStructRes{Result:"ok"}
	reqStruct :=new(proto.BStruct)
	err := reqStruct.XXX_Unmarshal(bMsg)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("agent BMsgHandle error request struct")
		resString.Result = err.Error()
		bM := make([]byte,0)
		res,err1 := resString.XXX_Marshal(bM,false)
		if err1 != nil{
			log.Error().Str("req",string(bM)).Msg("Marshal err")
		}
		return res,err
	}
	reqType := reqStruct.Type
	resString.Type = reqType
	var result []byte
	handle,err := server.GetFunc(reqType)
	if err == nil{
		result,err = handle(bMsg,session)
		if err != nil {
			resString.Result = err.Error()
		}else{
			resString.Content = result
		}
	}else{
		resString.Result = err.Error()
	}
	bM := make([]byte,0)
	bRes,err1 := resString.XXX_Marshal(bM,false)
	if err1 != nil{
		log.Error().Str("req",string(bM)).Msg("Marshal err")
	}

	agentMsg := proto.AgentReq{ReqType: "agent_pass", ReqContent:bMsg,ReqToken:base.RandToken()}
	tmpMsg := make([]byte,0)
	res,err := agentMsg.XXX_Marshal(tmpMsg,false)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("add json err")
	}else{
		log.Info().Str("req",string(res)).Msg("agent_pass send msg")
		err = server.SendServerMessage(res,"router",base.TypeRandom,0)
	}

	return bRes,err
}

func (server *AgentServer)RouterMsgHandle(bMsg []byte,session *base.Session) ([]byte,error){
	fmt.Println("RouterMsgHandle :" ,string(bMsg))
	reqStruct :=new(proto.AgentRes)
	err := reqStruct.XXX_Unmarshal(bMsg)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("agent BMsgHandle error request struct")
		return []byte{},err
	}
	log.Info().Str("receive",string(bMsg))

	return nil,err
}

// 发送socket消息
func (server *AgentServer)SendServerMessage(bMsg []byte,serverName string,rType int,key int64) error{
	sClient,err := server.GetSClient(bMsg,serverName,rType,key)
	if err != nil {
		err = server.TryConnect(serverName)
		if err != nil {
			log.Error().Str("connect","no connect").Send()
			return err
		}
		sClient,err = server.GetSClient(bMsg,serverName,rType,key)
		if err != nil {
			log.Error().Str("try_connect","try no connect").Send()
			return errors.New("no client")
		} else{
			log.Info().Str("tryConnect","has connect").Send()
		}
	}
	log.Info().Str("session_id",sClient.Session.SessionId()).Str("send",string(bMsg)).Msg("SendServerMessage start")
	sClient.Session.Write(bMsg)
	return nil
}
