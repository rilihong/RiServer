package routerserver

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"rilihong/RiServer/base"
	"rilihong/RiServer/proto"
)

type RouterServer struct {
	base.Server
}

func NewRouterServer(port int) *RouterServer{
	routerServer := new(RouterServer)
	routerServer.Init(port,"router",1930001)
	routerServer.SetMsgHandle(routerServer.BMsgHandle)
	routerServer.RegisterFunc("agent_pass",routerServer.AgentPass)
	return routerServer
}

func (routerServer *RouterServer)BMsgHandle(bMsg []byte,session *base.Session) ([]byte,error){
	reqStruct :=new(proto.AgentReq)
	err := reqStruct.XXX_Unmarshal(bMsg)
	if err != nil {
		log.Error().Str("req",string(bMsg)).Msg("BMsgHandle error request struct")
	}
	reqType := reqStruct.ReqType
	handle,err := routerServer.GetFunc(reqType)
	if err != nil{
		log.Error().Str("req",string(bMsg)).Str("type",reqType).Msg("error request type")
		return []byte{},errors.New("func not exit")
	}
	resString,err := handle(reqStruct.ReqContent,session)
	if err != nil{
		log.Error().Str("req",string(bMsg)).Str("type",reqType).Msg("error request type")
		return resString,nil
	} else {
		log.Info().Str("req",string(bMsg)).Str("res",string(resString)).Msg("res success")
		aRes := &proto.AgentRes{ResType:"agent_receive",ResContent:resString,ReqToken:reqStruct.ReqToken}
		aMsg := make([]byte,0)
		re,_ := aRes.XXX_Marshal(aMsg,false)
		return re,nil
	}
}

func (routerServer *RouterServer)AgentPass(bMsg []byte,session *base.Session) ([]byte,error){
	fmt.Println(string(bMsg))
	bStructMsg := new(proto.BStruct)
	bStructMsg.XXX_Unmarshal(bMsg)

	bStructRes := new(proto.BStructRes)
	if bStructMsg.Type == "login" {
		bStructRes.Type = "loginRes"
		bStructRes.Result = "ok"
		bStructRes.Content = []byte("routerBack")
	}else{
		bStructRes.Type = "unKnown"
		bStructRes.Result = "err"
		bStructRes.Content = []byte("seek routerBack")
	}
	tmpMsg := make([]byte,0)
	ls,_ := bStructRes.XXX_Marshal(tmpMsg,false)
	fmt.Println("end send ",string(ls))
	return ls,nil
}