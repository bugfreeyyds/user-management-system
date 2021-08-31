package tcpserver

import (
	"context"

	"ums/code"
	pb "ums/proto"
	"ums/utils"

	logs "github.com/beego/beego/adapter/logs"
	"google.golang.org/grpc/metadata"
)

// UserServer for rcpclient
type UserServer struct {
	API *API
}

func getUUID(ctx context.Context) string {
	var uuid string
	md, ok := metadata.FromIncomingContext(ctx)
	if ok == false {
		return uuid
	}
	uuids := md.Get("uuid")
	if len(uuids) == 1 {
		uuid = uuids[0]
	}
	return uuid
}

// Login login handler
func (s *UserServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	// get uuid
	uuid := getUUID(ctx)
	logs.Debug(uuid, " -- Login access from:", in.Username, "@", in.Passwd)
	// query userinfo
	user, err := getUserInfo(in.Username)
	if err != nil {
		logs.Error(uuid, " -- Failed to getUserInfo, ", in.Username, "@", in.Passwd, ", err:", err.Error())
		return &pb.LoginResponse{Code: code.CodeTCPFailedGetUserInfo, Msg: code.CodeMsg[code.CodeTCPFailedGetUserInfo]}, nil
	}

	// verify passwd
	if utils.Md5String(in.Passwd+user.Skey) != user.Passwd {
		logs.Error(uuid, " -- Failed to match passwd ", in.Username, "@", in.Passwd, " salt:", user.Skey, " realpwd:", user.Passwd)
		return &pb.LoginResponse{Code: code.CodeTCPPasswdErr, Msg: code.CodeMsg[code.CodeTCPPasswdErr]}, nil
	}

	// set cache
	token := utils.GenerateToken(user.Username)
	err = s.API.redisClient.SetTokenInfo(user, token)
	if err != nil {
		logs.Error(uuid, " -- Failed to set token for user:", user.Username, " err:", err.Error())
		return &pb.LoginResponse{Code: code.CodeTCPInternelErr, Msg: code.CodeMsg[code.CodeTCPInternelErr]}, nil
	}
	logs.Debug(uuid, " -- Login succesfully, ", in.Username, "@", in.Passwd, " with token:", token)
	return &pb.LoginResponse{Username: user.Username, Nickname: user.Nickname, Headurl: user.Headurl, Token: token, Code: code.CodeSucc}, nil
}

// GetUserInfo get user info
func (s *UserServer) GetUserInfo(ctx context.Context, in *pb.CommRequest) (*pb.LoginResponse, error) {
	// get uuid
	uuid := getUUID(ctx)
	logs.Debug(uuid, " -- GetUserInfo access from:", in.Username, " with token:", in.Token)
	// get and verify token
	token := in.Token
	if len(token) != 32 {
		logs.Error(uuid, " -- Error: invalid token:", in.Token)
		return &pb.LoginResponse{Code: code.CodeTCPInvalidToken, Msg: code.CodeMsg[code.CodeTCPInvalidToken]}, nil
	}
	// get userinfo and compare username
	user, err := s.API.redisClient.GetTokenInfo(token)
	if err != nil {
		logs.Error(uuid, " -- Failed to get token:", in.Token, " with err:", err.Error())
		return &pb.LoginResponse{Code: code.CodeTCPTokenExpired, Msg: code.CodeMsg[code.CodeTCPTokenExpired]}, nil
	}

	// check if username is the same
	if user.Username != in.Username {
		logs.Error(uuid, " -- Error: token info not match:", in.Username, " while cache:", user.Username)
		return &pb.LoginResponse{Code: code.CodeTCPUserInfoNotMatch, Msg: code.CodeMsg[code.CodeTCPUserInfoNotMatch]}, nil
	}
	logs.Debug(uuid, " -- Succ to GetUserInfo :", in.Username, " with token:", in.Token)
	return &pb.LoginResponse{Username: user.Username, Nickname: user.Nickname, Headurl: user.Headurl, Token: token, Code: code.CodeSucc}, nil
}

// EditUserInfo edit userinfo (nickname, headurl or both)
func (s *UserServer) EditUserInfo(ctx context.Context, in *pb.EditRequest) (*pb.EditResponse, error) {
	// get uuid
	uuid := getUUID(ctx)
	logs.Debug(uuid, " -- EditUserInfo access from:", in.Username, " with token:", in.Token)
	// auth
	authResult := s.API.Auth(in.Username, in.Token)
	if authResult == false {
		logs.Error(uuid, " -- Failed to auth for user:", in.Username, " with token:", in.Token)
		return &pb.EditResponse{Code: code.CodeTCPTokenExpired, Msg: code.CodeMsg[code.CodeTCPTokenExpired]}, nil
	}
	affectRows := s.API.EditUserInfo(in.Username, in.Nickname, in.Headurl, in.Token, in.Mode)
	logs.Error(uuid, " -- Succ to edit userinfo, affected rows is:", affectRows)
	return &pb.EditResponse{Code: code.CodeSucc, Msg: code.CodeMsg[code.CodeSucc]}, nil
}

// Logout logout
func (s *UserServer) Logout(ctx context.Context, in *pb.CommRequest) (*pb.EditResponse, error) {
	// get uuid
	uuid := getUUID(ctx)
	logs.Debug(uuid, " -- Logout access from:", in.Token)
	err := s.API.redisClient.DelTokenInfo(in.Token)
	if err != nil {
		logs.Error(uuid, " -- Failed to delTokenInfo :", err.Error())
	}
	logs.Debug(uuid, " -- Succ to logout:", in.Token)
	return &pb.EditResponse{Code: code.CodeSucc, Msg: code.CodeMsg[code.CodeSucc]}, nil
}
