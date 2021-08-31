package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redis "github.com/go-redis/redis/v8"
	"ums/conf"
	"ums/tcpserver/consts"
	"ums/tcpserver/types"
)

type cacheConfig struct {
	tokenExpired int
	userExpired  int
}

type RedisClient struct {
	client      *redis.Client
	cacheConfig *cacheConfig
}

// init redis connection pool
func NewRedisClient(conf *conf.TCPConf) (*RedisClient, error) {
	redisConn := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,
		Password: conf.Redis.Passwd,
		DB:       conf.Redis.Db,
		PoolSize: conf.Redis.Poolsize,
	})
	if redisConn == nil {
		return nil, errors.New("Failed to call redis.NewClient")
	}

	_, err := redisConn.Ping(context.Background()).Result()
	if err != nil {
		msg := fmt.Sprintf("Failed to ping redis, err:%s", err.Error())
		return nil, errors.New(msg)
	}

	client := &RedisClient{
		client: redisConn,
		cacheConfig: &cacheConfig{
			tokenExpired: conf.Redis.Cache.Tokenexpired,
			userExpired:  conf.Redis.Cache.Userexpired,
		},
	}
	return client, nil
}

// cleanup
func (c *RedisClient) CloseCache() error {
	return c.client.Close()
}

// get cached userinfo
func (c *RedisClient) GetUserCacheInfo(username string) (types.User, error) {
	redisKey := consts.UserInfoPrefix + username
	val, err := c.client.Get(context.Background(), redisKey).Result()
	var user types.User
	if err != nil {
		return user, err
	}
	err = json.Unmarshal([]byte(val), &user)
	return user, err
}

// set cached userinfo
func (c *RedisClient) SetUserCacheInfo(user types.User) error {
	redisKey := consts.UserInfoPrefix + user.Username
	val, err := json.Marshal(user)
	if err != nil {
		return err
	}
	expired := time.Second * time.Duration(c.cacheConfig.userExpired)
	_, err = c.client.Set(context.Background(), redisKey, val, expired).Result()
	return err
}

// get token info
func (c *RedisClient) GetTokenInfo(token string) (types.User, error) {
	redisKey := consts.TokenKeyPrefix + token
	val, err := c.client.Get(context.Background(), redisKey).Result()
	var user types.User
	if err != nil {
		return user, err
	}
	err = json.Unmarshal([]byte(val), &user)
	return user, err
}

// set cached userinfo
func (c *RedisClient) SetTokenInfo(user types.User, token string) error {
	redisKey := consts.TokenKeyPrefix + token
	val, err := json.Marshal(user)
	if err != nil {
		return err
	}
	expired := time.Second * time.Duration(c.cacheConfig.tokenExpired)
	_, err = c.client.Set(context.Background(), redisKey, val, expired).Result()
	return err
}

// update cached userinfo, if failed, try to delete it from cache
func (c *RedisClient) UpdateCachedUserinfo(user types.User) error {
	err := c.SetUserCacheInfo(user)
	if err != nil {
		redisKey := consts.UserInfoPrefix + user.Username
		c.client.Del(context.Background(), redisKey).Result()
	}
	return err
}

// delete token cache info
func (c *RedisClient) DelTokenInfo(token string) error {
	redisKey := consts.TokenKeyPrefix + token
	_, err := c.client.Del(context.Background(), redisKey).Result()
	return err
}
