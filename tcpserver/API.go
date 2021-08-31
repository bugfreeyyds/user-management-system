package tcpserver

import (
	"os"

	"ums/conf"
	"ums/tcpserver/cache"
	"ums/tcpserver/consts"
	"ums/tcpserver/db"
	"ums/tcpserver/types"
)

// API
type API struct {
	redisClient *cache.RedisClient
	dbClient    *db.DBClient
}

// NewAPI new a API
func NewAPI(config *conf.TCPConf) *API {
	// init redis
	redisClient, err := cache.NewRedisClient(config)
	if err != nil {
		//logs.Critical("initRedisConn failed:", err.Error())
		os.Exit(-1)
	}

	// init db
	dbClient, err := db.NewDBClient(config)
	if err != nil {
		//logs.Critical("initDbConn failed:", err.Error())
		os.Exit(-1)
	}
	//logs.Info("init successfully!")
	return &API{
		redisClient: redisClient,
		dbClient:    dbClient,
	}
}

// Finalize clean up the cache and db resources
func (a *API) Finalize() {
	a.redisClient.CloseCache()
	a.dbClient.CloseDB()
}

// GetUserInfo get user info
func (a *API) GetUserInfo(username string) (types.User, error) {
	// try cache
	user, err := a.redisClient.GetUserCacheInfo(username)
	if err == nil && user.Username == username {
		return user, err
	}

	// get from db
	user, err = a.dbClient.GetDbUserInfo(username)
	if err != nil {
		return user, err
	}

	// update cache
	serr := a.redisClient.SetUserCacheInfo(user)
	if serr != nil {
		//logs.Error("cache userinfo failed for user:", user.Username, " with err:", serr.Error())
	}

	return user, err
}

// EditUserInfo edit user info
func (a *API) EditUserInfo(username, nickname, headurl, token string, mode uint32) int64 {
	// update db info
	var affectedRows int64
	switch mode {
	case consts.EditUsername:
		affectedRows = a.dbClient.UpdateDbNickname(username, nickname)
	case consts.EditHeadurl:
		affectedRows = a.dbClient.UpdateDbHeadurl(username, headurl)
	case consts.EditBoth:
		affectedRows = a.dbClient.UpdateDbUserinfo(username, nickname, headurl)
	default:
		// do nothing
		break
	}

	// on successing, update cache or delete it if updating failed
	if affectedRows == 1 {
		user, err := a.dbClient.GetDbUserInfo(username)
		if err == nil {
			a.redisClient.UpdateCachedUserinfo(user)
			if token != "" {
				err = a.redisClient.SetTokenInfo(user, token)
				if err != nil {
					//logs.Error("update token failed:", err.Error())
					a.redisClient.DelTokenInfo(token)
				}
			}
		} else {
			//logs.Error("Failed to get dbUserInfo for cache, username:", username, " with err:", err.Error())
		}
	}
	return affectedRows
}

// Auth authenticate username
func (a *API) Auth(username, token string) bool {
	user, err := a.redisClient.GetTokenInfo(token)
	if err != nil {
		//logs.Error("failed to getTokenInfo, token:", token)
		return false
	}
	if user.Username != username {
		//logs.Error("invalid token info, username not match!")
		return false
	}
	return true
}
