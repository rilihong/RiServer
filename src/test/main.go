package main

import (
	"fmt"
	"log"
	"net"
	"rilihong/RiServer/proto"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:1920")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	//time.Sleep(time.Second * 5)
	bStruct := proto.BStruct{Type:"login",Content:[]byte("sdadad")}
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
	fmt.Println(string(cc))
}