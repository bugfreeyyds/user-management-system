package rpcclient

import (
    "context"
    "net/http"
    "strconv"
    "time"

    gpool "user-management-system/httpserver/rpcclient/gpool"
    "user-management-system/type/code"
    pb "user-management-system/type/proto"

    "github.com/gin-gonic/gin"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
    log "github.com/beego/beego/v2/adapter/logs"
)

var pool *gpool.GPool

// InitPool  init grpc client connection pool
func InitPool(addr string, init, capacity uint32, maxIdle time.Duration) error {
    // init grpc client pool
    var err error
    pool, err = gpool.NewPool(func () (*grpc.ClientConn, error) {
                     conn, err := grpc.Dial(addr, grpc.WithInsecure())
                     if err != nil {
                         return nil, err
                     }
                     return conn, nil
                 },
                 init, capacity, maxIdle)
    return err
}

// DestoryPool destroy connection pool
func DestoryPool() {
    pool.Close()
}

// clientWrap
type clientWrap struct {
    conn   *gpool.Conn
    client pb.UserServiceClient
}

// getRPCClient get a rpc client
func getRPCClient() (*clientWrap, error) {
    // get conn
    ctx, cancel := context.WithDeadline(context.Background(),  time.Now().Add(10 * time.Millisecond))
    conn, err := pool.Get(ctx)
    cancel()
    if err != nil {
        return nil, err
    }

    client := pb.NewUserServiceClient(conn.C)
    return &clientWrap{conn, client}, nil
}

// freeRPCClient free a rpc client
func freeRPCClient(wrap* clientWrap) {
    err := pool.Put(wrap.conn)
    if err != nil {
        log.Error("Failed to reclaime conn, err:", err.Error())
    }
}

// FormatResponse : suppress golint error
/* {
 *   c: int   // error code, 0 for succ
 *   msg: string // succ or error msg
 *   data: {     // response data object
 *   }
 * }
 */
func FormatResponse(c int, msg string, data map[string]string) map[string]interface{} {
    if msg == "" {
        msg = code.CodeMsg[c]
    }

    return gin.H{"code": c, "msg": msg, "data": data}
}

// Login : userlogin handler
func Login(args map[string]string) (int, string, map[string]interface{}) {
    // get uuid
    uuid := args["uuid"]
    // communicate with rcp server
    client, err := getRPCClient()
    if err != nil {
        log.Error(uuid, " -- Failed to getRPCClient, err:", err.Error())
        return http.StatusInternalServerError, "", FormatResponse(code.CodeInternalErr, "", nil)
    }
    defer freeRPCClient(client)

    ctx := metadata.AppendToOutgoingContext(context.Background(), "uuid", uuid)
    rsp, err := client.client.Login(ctx, &pb.LoginRequest{Username: args["username"], Passwd: args["passwd"]})
    if err != nil {
        log.Error(uuid, " -- Failed to communicate with TCP server, err:", err.Error())
        return http.StatusOK, "", FormatResponse(code.CodeErrBackend, "", nil)
    }

    log.Debug(uuid, " -- Succ get token:", rsp.Token, " code:", rsp.Code)

    var token string
    if rsp.Code == code.CodeSucc && rsp.Token != "" {
        token = rsp.Token
    }

    return http.StatusOK, token, FormatResponse(int(rsp.Code), rsp.Msg, map[string]string{"username":rsp.Username, "nickname":rsp.Nickname, "headurl":rsp.Headurl})
}

// Logout : user logout
func Logout(args map[string]string) (int, map[string]interface{}) {
    // get uuid
    uuid := args["uuid"]
    // communicate with rcp server
    client, err := getRPCClient()
    if err != nil {
        log.Error(uuid, " -- Failed to getRPCClient, err:", err.Error())
        return http.StatusInternalServerError, FormatResponse(code.CodeInternalErr, "", nil)
    }
    defer freeRPCClient(client)

    ctx := metadata.AppendToOutgoingContext(context.Background(), "uuid", uuid)
    rsp, err := client.client.Logout(ctx, &pb.CommRequest{Token: args["token"], Username: args["username"]})
    if err != nil {
        log.Error(uuid, " -- Failed to communicate with TCP server, err:", err.Error())
        return http.StatusOK, FormatResponse(code.CodeErrBackend, "", nil)
    }
    log.Debug(uuid, "Succ to get response from backend with ", rsp.Code, " and msg:", rsp.Msg)

    return http.StatusOK, FormatResponse(int(rsp.Code), rsp.Msg, nil)
}

// EditUserinfo  edit user nickname/headurl
func EditUserinfo(args map[string]string) (int, map[string]interface{}) {
    // get uuid
    uuid := args["uuid"]

    headurl := args["headurl"]
    // get connection
    client, err := getRPCClient()
    if err != nil {
        log.Error(uuid, " -- Failed to getRPCClient, err:", err.Error())
        return http.StatusInternalServerError, FormatResponse(code.CodeInternalErr, "", nil)
    }
    defer freeRPCClient(client)

    // update userinfo
    mode, _ := strconv.Atoi(args["mode"])
    ctx := metadata.AppendToOutgoingContext(context.Background(), "uuid", uuid)
    editRsp, err := client.client.EditUserInfo(ctx,
                          &pb.EditRequest{Username: args["username"], Token: args["token"], Nickname: args["nickname"], Headurl: headurl, Mode: uint32(mode)})
    if err != nil {
        log.Error(uuid, " -- Failed to communicate with TCP server, err:", err.Error())
        return http.StatusOK, FormatResponse(code.CodeErrBackend, "", nil)
    }
    data := map[string]string{}
    if editRsp.Code == 0 && headurl != "" {
        data["headurl"] = headurl
    }

    return http.StatusOK, FormatResponse(int(editRsp.Code), editRsp.Msg, data)
}

// GetUserinfo get userinfo handler
func GetUserinfo(args map[string]string) (int, map[string]interface{}) {
    // get uuid
    uuid := args["uuid"]
    // communicate with rcp server
    client, err := getRPCClient()
    if err != nil {
        log.Error(uuid, " -- Failed to getRPCClient, err:", err.Error())
        return http.StatusInternalServerError, FormatResponse(code.CodeInternalErr, "", nil)
    }
    defer freeRPCClient(client)

    ctx := metadata.AppendToOutgoingContext(context.Background(), "uuid", uuid)
    rsp, err := client.client.GetUserInfo(ctx, &pb.CommRequest{Token: args["token"], Username: args["username"]})
    if err != nil {
        log.Error(uuid, " -- Failed to communicate with TCP server, err:", err.Error())
        return http.StatusOK, FormatResponse(code.CodeErrBackend, "", nil)
    }
    response := FormatResponse(int(rsp.Code), rsp.Msg, map[string]string{"username":rsp.Username, "nickname":rsp.Nickname, "headurl":rsp.Headurl})

    return http.StatusOK, response
}

// Auth user getUserInfo to
func Auth(args map[string]string) (int, int, string) {
    // get uuid
    uuid := args["uuid"]
    // communicate with rcp server
    client, err := getRPCClient()
    if err != nil {
        log.Error(uuid, " -- Failed to getRPCClient, err:", err.Error())
        return http.StatusInternalServerError, code.CodeInternalErr, code.CodeMsg[code.CodeInternalErr]
    }
    defer freeRPCClient(client)

    ctx := metadata.AppendToOutgoingContext(context.Background(), "uuid", uuid)
    rsp, err := client.client.GetUserInfo(ctx, &pb.CommRequest{Token: args["token"], Username: args["username"]})
    if err != nil {
        log.Error(uuid, " -- Failed to communicate with TCP server, err:", err.Error())
        return http.StatusOK, code.CodeErrBackend, code.CodeMsg[code.CodeErrBackend]
    }
    if rsp.Code == 0 {
        return http.StatusOK, code.CodeSucc, code.CodeMsg[code.CodeSucc]
    }

    return http.StatusOK, int(rsp.Code), rsp.Msg
}
