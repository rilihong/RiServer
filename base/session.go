package base

import (
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"sync"
)

type Session struct{
	conn net.Conn
	isLive bool
	sessionId string
	buffer chan []byte
	handle MsgHandle
	sync.RWMutex
}

func NewSession(conn net.Conn ,handle MsgHandle) *Session{
	session := new(Session)
	session.conn = conn
	session.isLive = true
	session.buffer = make(chan []byte,1024)
	session.sessionId = SessionId()
	session.handle = handle
	return session
}

func (session *Session)SessionId() string{
	return session.sessionId
}

func (session *Session)Write(out []byte){
	if session.isLive == false{
		return
	}
	session.buffer <- out
}

func (session *Session)WriteMsg(msg string){
	session.Write([]byte(msg))
}

func (session *Session)SessionWrite(){
	for{
		if session.isLive == false{
			return
		}
		msg := <-session.buffer
		if msg == nil{
			log.Error().Str("session id",session.sessionId).Msg("session error close")
			return
		}
		session.conn.Write(msg)
		log.Info().Str("session id",session.sessionId).Str("content",string(msg)).Send()
	}
}
func (session *Session)ReadRequest() ([]byte,error){
	tmpbuf := make([]byte,1024)
	len,err := session.conn.Read(tmpbuf)
	if err != nil && err != io.EOF {
		return nil,err
	}
	var reqbuf = make([]byte,len)
	copy(reqbuf,tmpbuf)
	return reqbuf,err
}

func (session *Session)SessionRead(){
	for{
		if session.isLive == false{
			return
		}
		req,err := session.ReadRequest()
		if err != nil {
			session.Close()
			return
		}
		res,err := session.handle(req, session)
		if res != nil{
			session.Write(res)
		}
	}
}

func (session *Session)Close(){
	session.Lock()
	if session.isLive == false{
		session.Unlock()
		return
	}
	session.conn.Close()
	session.isLive = false
	session.Unlock()
	close(session.buffer)
}

func (session *Session)GetConn() net.Conn{
	return session.conn
}