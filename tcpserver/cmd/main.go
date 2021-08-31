package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	"ums/conf"
	pb "ums/proto"
	"ums/tcpserver"

	logs "github.com/beego/beego/adapter/logs"
	"google.golang.org/grpc"
)

// run starts UserServer services
func run(config *conf.TCPConf, api *tcpserver.API) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Server.Port))
	if err != nil {
		logs.Critical("Listen failed, err:", err.Error())
		return
	}

	userServer := &tcpserver.UserServer{API: api}
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, userServer)

	logs.Info("start to listen on localhost:%d", config.Server.Port)
	err = grpcServer.Serve(lis)
	if err != nil {
		fmt.Println("Server failed, err:", err.Error())
	}
}

func main() {
	// init log
	//logConfig := fmt.Sprintf(`{"filename":"%s","level":%s,"maxlines":0,"maxsize":0,"daily":true,"maxdays":%s}`,
	//                        config.Log.Logfile, config.Log.Loglevel, config.Log.Maxdays)
	//logs.SetLogger(logs.AdapterFile, logConfig)
	//logs.EnableFuncCallDepth(true)
	//logs.SetLogFuncCallDepth(3)
	//logs.Async()

	var config conf.TCPConf
	// parser config
	var confFile string
	flag.StringVar(&confFile, "c", "../conf/tcpserver.yaml", "config file")
	flag.Parse()

	err := conf.ConfParser(confFile, &config)
	if err != nil {
		//logs.Critical("parser config failed:", err.Error())
		os.Exit(-1)
	}

	aAPI := tcpserver.NewAPI(&config)
	defer aAPI.Finalize()

	// generate random seed global
	rand.Seed(time.Now().UTC().UnixNano())
	// start event loop
	run(&config, aAPI)
}
