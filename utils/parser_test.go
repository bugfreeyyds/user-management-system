package utils

import (
    "fmt"
    "testing"
    "user-management-system/conf"
)

func Test_ConfParser(t *testing.T) {
    // http
    var httpConf conf.HTTPConf
    err := ConfParser("../conf/httpserver.yaml", &httpConf)
    if err != nil {
        t.Error(err.Error())
    }
    fmt.Println(httpConf.Server.IP, httpConf.Server.Port)

    // tcp
    var tcpConf conf.TCPConf
    err = ConfParser("../conf/tcpserver.yaml", &tcpConf)
    if err != nil {
        t.Error(err.Error())
    }
}
