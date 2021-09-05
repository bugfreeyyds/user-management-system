package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"user-management-system/conf"
	"user-management-system/httpserver/rpcclient"
	"user-management-system/utils"

	log "github.com/beego/beego/v2/adapter/logs"
	"github.com/gin-gonic/gin"
)

var config conf.HTTPConf

// init parse config and initialize log and rpc connection pool
func init() {
	// parser config
	var confFile string
	flag.StringVar(&confFile, "c", "./conf/httpserver.yaml", "config file")
	flag.Parse()

	err := utils.ConfParser(confFile, &config)
	if err != nil {
		log.Critical("Parser config failed, err:", err.Error())
		os.Exit(-1)
	}
    log.Info("parse config successfully!")

	//init log
	logConfig := fmt.Sprintf(`{"filename":"%s","level":%s,"maxlines":0,"maxsize":0,"daily":true,"maxdays":%s}`,
		config.Log.Logfile, config.Log.Loglevel, config.Log.Maxdays)
	log.SetLogger(log.AdapterFile, logConfig)
	log.EnableFuncCallDepth(true)
	log.SetLogFuncCallDepth(3)
	log.Async()

	// init rpcclient pool
	err = rpcclient.InitPool(config.Rpcserver.Addr, config.Pool.Initsize, config.Pool.Capacity, time.Duration(config.Pool.Maxidle)*time.Second)
	if err != nil {
		log.Critical("InitPool failed, err:", err.Error())
		os.Exit(-2)
	}

    log.Info("httpserver init finished: listen on port: %d", config.Server.Port)
}

// finalize destroy rpcclient pool
func finalize() {
	rpcclient.DestoryPool()
}

func main() {
	defer finalize()

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = ioutil.Discard

	engine := gin.Default()
	engine.Any("/welcome", webRoot)
	engine.POST("/login", loginHandler)
	engine.POST("/logout", logoutHandler)
	engine.GET("/getuserinfo", getUserinfoHandler)
	engine.POST("/editnickname", editNicknameHandler)
	engine.POST("/uploadpic", uploadHeadurlHandler)

	engine.POST("/randlogin", randomLoginHandler)
	engine.Static("/static/", "./static/")
	engine.Static("/upload/images/", "./upload/images/")

	engine.Run(fmt.Sprintf(":%d", config.Server.Port))
}

func webRoot(context *gin.Context) {
	context.String(http.StatusOK, "hello, world")
}
