package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"rilihong/RiServer/src/router/routerserver"
	"strconv"
)

func main(){
	logPath := string("./server_") + strconv.Itoa(os.Getpid()) + ".log"
	file,err := os.OpenFile(logPath,os.O_RDWR|os.O_CREATE|os.O_APPEND,755)
	if err != nil{
		fmt.Println("add log err")
	}
	log.Logger = log.Output(file)
	server := routerserver.NewRouterServer(1930)
	server.ListenAndServe()
}