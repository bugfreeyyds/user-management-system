package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"

	"user-management-system/conf"
	"user-management-system/utils"
	pb "user-management-system/type/proto"
	"user-management-system/tcpserver"

	log "github.com/beego/beego/v2/adapter/logs"
	"google.golang.org/grpc"
)

// run starts UserServer services
func run(config *conf.TCPConf, api *tcpserver.API) {
	userServer := &tcpserver.UserServer{API: api}
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, userServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Server.Port))
	if err != nil {
		log.Critical("Listen failed, err:", err.Error())
		return
	}

	log.Info("start to listen on localhost:%d", config.Server.Port)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Critical("Server failed, err:", err.Error())
        os.Exit(-1)
	}
}

func main() {
	// parser config
	var config conf.TCPConf
	var confFile string
	flag.StringVar(&confFile, "c", "./conf/tcpserver.yaml", "config file")
	flag.Parse()

	err := utils.ConfParser(confFile, &config)
	if err != nil {
        log.Critical("parse config failed: error: %v\n", err)
		os.Exit(-1)
	}
    log.Info("parse config successfully!")

	// init log
	logConfig := fmt.Sprintf(`{"filename":"%s","level":%s,"maxlines":0,"maxsize":0,"daily":true,"maxdays":%s}`,
	                        config.Log.Logfile, config.Log.Loglevel, config.Log.Maxdays)
	log.SetLogger(log.AdapterFile, logConfig)
	log.EnableFuncCallDepth(true)
	log.SetLogFuncCallDepth(3)
	log.Async()
    log.Info("init log finished!")

	aAPI := tcpserver.NewAPI(&config)
	defer aAPI.Finalize()
    log.Debug("new API successfully, new api: %v\n", aAPI)

	// generate random seed global
	rand.Seed(time.Now().UTC().UnixNano())
	// start event loop
	run(&config, aAPI)
}
