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
	TokenMap map[string]*base.Session
}

func NewAgentServer(port int) *AgentServer{
	agentServer := new(AgentServer)
	agentServer.Init(port,"agent",1920001)
	agentServer.TokenMap = make(map[string]*base.Session)
	agentServer.SetMsgHandle(agentServer.BMsgHandle)
	agentServer.SetRouterHandle(agentServer.RouterMsgHandle)
	agentServer.RegisterFunc("login",agentServer.NormalHandle)
	agentServer.RegisterFunc("loginRes",agentServer.RouterResHandle)
	return agentServer
}

func (server *AgentServer)NormalHandle(bMsg []byte,session *base.Session) ([]byte,string,error){
	resType := "loginRes"
	reqStruct :=new(proto.BStruct)
	err := reqStruct.XXX_Unmarshal(bMsg)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("NormalHandle error request struct")
		return nil,resType,err
	}
	fmt.Println("type:",reqStruct.GetType()," content:",string(reqStruct.GetContent()))
	bLen := len(reqStruct.GetContent())
	res := make([]byte,bLen)
	copy(res,reqStruct.GetContent())
	return res,resType,nil
}

func (server *AgentServer)RouterResHandle(bMsg []byte,session *base.Session) ([]byte,string,error){

	agentRes :=new(proto.AgentRes)
	err := agentRes.XXX_Unmarshal(bMsg)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("RouterHandle error request struct")
		return nil,"",err
	}
	fmt.Println("type:",agentRes.GetResType()," content:",string(agentRes.GetResContent()))
	bLen := len(agentRes.GetResContent())
	res := make([]byte,bLen)
	copy(res,agentRes.GetResContent())
	return res,agentRes.ResType,nil
}

func (server *AgentServer)BMsgHandle(bMsg []byte,session *base.Session) ([]byte,error){
	reqStruct :=new(proto.BStruct)
	//解析异常，不返回
	err := reqStruct.XXX_Unmarshal(bMsg)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Str("err",err.Error()).Msg("agent BMsgHandle error request struct")
		return nil,err
	}
	//有处理函数，且函数处理异常，返回异常信息
	resString := &proto.BStructRes{Result:"ok"}
	handle,err := server.GetFunc(reqStruct.Type)
	if err == nil{
		result,sType,err := handle(bMsg,session)
		if err != nil {
			resString.Type = sType
			resString.Result = err.Error()
			resString.Content = result
			bM := make([]byte,0)
			res,_ := resString.XXX_Marshal(bM,false)
			return res,err
		}
	}
	//透传到router
	agentMsg := proto.AgentReq{ReqType: reqStruct.Type, ReqContent:reqStruct.Content,ReqToken:base.RandToken()}
	server.TokenMap[agentMsg.ReqToken] = session
	tmpMsg := make([]byte,0)
	res,err := agentMsg.XXX_Marshal(tmpMsg,false)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("add json err")
	}else{
		log.Info().Str("req",string(res)).Msg("agent_pass send msg")
		err = server.SendServerMessage(res,"router",base.TypeRandom,0)
	}

	return nil,err
}

func (server *AgentServer)RouterMsgHandle(bMsg []byte,session *base.Session) ([]byte,error){
	fmt.Println("RouterMsgHandle :" ,string(bMsg))
	resStruct :=new(proto.AgentRes)
	err := resStruct.XXX_Unmarshal(bMsg)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("agent BMsgHandle error request struct")
		return []byte{},err
	}
	token := resStruct.ReqToken
	link,ok := server.TokenMap[token]
	if ok {
		res := &proto.BStructRes{Type:resStruct.ResType,Result:"ok",Content:resStruct.ResContent}
		tmpMsg := make([]byte,0)
		rMsg,err := res.XXX_Marshal(tmpMsg,false)
		if err == nil {
			link.Write(rMsg)
		}
	}

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
