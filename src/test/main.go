package main

import (
	"fmt"
	"log"
	"net"
	"rilihong/RiServer/proto"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:1920")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	//time.Sleep(time.Second * 5)
	tick := time.NewTicker(time.Second * 2)
	for{
		select {
			case <- tick.C:
				sendMsg(conn)
		}
	}
}

func sendMsg(conn net.Conn){
	cont := &proto.PokerReq{Uid:1024,Name:"kitty"}
	bb := make([]byte,0)
	bCon,err := cont.XXX_Marshal(bb,false)
	if err != nil {
		fmt.Println(err.Error())
	}
	bStruct := proto.BStruct{Type:"login",Content:bCon}
	bMsg := make([]byte,0)
	ls,err := bStruct.XXX_Marshal(bMsg,true)
	if err != nil {
		fmt.Println(err.Error())
	}
	str := string(ls)
	fmt.Println(str," len ",len(ls))

	conn.Write(ls)

	var ans = make([]byte,1024)
	bLen,_ := conn.Read(ans)
	cc := make([]byte,bLen)
	copy(cc,ans)
	res := new(proto.BStructRes)
	err = res.XXX_Unmarshal(cc)
	if err != nil {
		fmt.Println("respond err")
	}
	fmt.Println(res)
}