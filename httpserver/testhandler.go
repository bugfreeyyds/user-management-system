package main

import (
    "fmt"
    "math/rand"
    "net/http"

    "user-management-system/httpserver/rpcclient"
    "user-management-system/type/code"

    "github.com/gin-gonic/gin"
    logs2 "github.com/beego/beego/v2/adapter/logs"
)

// login
func randomLoginHandler(c *gin.Context) {
    // check params
    uid := rand.Int63n(10000000)
    username := fmt.Sprintf("username%d", uid)
    passwd := "e10adc3949ba59abbe56e057f20f883e"

    if len(passwd) != 32 {
        logs2.Error("Invalid passwd:", passwd)
        c.JSON(http.StatusBadRequest, rpcclient.FormatResponse(code.CodeInvalidPasswd, "", nil))
        return
    }

    // communicate with rcp server
    ret, token, rsp := rpcclient.Login(map[string]string{"username":username, "passwd":passwd})
    // set cookie
    logs2.Debug("set cookie with expire:", token)
    if ret == http.StatusOK && token != "" {
        c.SetCookie("token", token, config.Logic.Tokenexpire, "/", config.Server.IP, false, true)
        logs2.Debug("set cookie with expire:", config.Logic.Tokenexpire)
    }

    logs2.Debug("succ get response from backend with", rsp["code"], " and msg:", rsp["msg"])
    c.JSON(ret, rsp)
}
