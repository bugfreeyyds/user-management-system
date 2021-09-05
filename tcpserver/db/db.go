package db

import (
	"errors"
	"fmt"
	"time"

	"user-management-system/conf"
	"user-management-system/tcpserver/types"
	"user-management-system/utils"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type DBClient struct {
	client *gorm.DB
}

// init conn
func NewDBClient(config *conf.TCPConf) (*DBClient, error) {
	conninfo := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", config.Db.User, config.Db.Passwd, config.Db.Host, config.Db.Db)
	client, err := gorm.Open("mysql", conninfo)
	if err != nil {
		msg := fmt.Sprintf("Failed to connect to db '%s', err: %s", conninfo, err.Error())
		return nil, errors.New(msg)
	}

	client.DB().SetMaxIdleConns(config.Db.Conn.Maxidle)
	client.DB().SetMaxOpenConns(config.Db.Conn.Maxopen)
	//db.LogMode(true)

	dbClient := &DBClient{client: client}
	return dbClient, nil
}

// cleanup
func (d *DBClient) CloseDB() error {
	return d.client.Close()
}

// query
func (d *DBClient) GetDbUserInfo(username string) (types.User, error) {
	var quser types.User
	d.client.Table(utils.GetTableName(username)).Where(&types.User{Username: username}).First(&quser)
	if quser.Username == "" {
		return quser, fmt.Errorf("user(%s) not exists", username)
	}
	return quser, nil
}

// update nickname
func (d *DBClient) UpdateDbNickname(username, nickname string) int64 {
	return d.client.Table(utils.GetTableName(username)).Model(&types.User{}).Where("`username` = ?", username).Updates(types.User{Nickname: nickname, Uptime: time.Now().Unix()}).RowsAffected
}

// update headurl
func (d *DBClient) UpdateDbHeadurl(username, url string) int64 {
	return d.client.Table(utils.GetTableName(username)).Model(&types.User{}).Where("`username` = ?", username).Updates(types.User{Headurl: url, Uptime: time.Now().Unix()}).RowsAffected
}

// update nickname and headurl
func (d *DBClient) UpdateDbUserinfo(username, nickname, url string) int64 {
	return d.client.Table(utils.GetTableName(username)).Model(&types.User{}).Where("`username` = ?", username).Updates(types.User{Nickname: nickname, Headurl: url, Uptime: time.Now().Unix()}).RowsAffected
}
