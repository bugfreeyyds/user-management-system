package types

import "fmt"

// User gorm user object
type User struct {
	ID       int32       `gorm:"types:int(11);primary key"`
	Username string      `gorm:"types:varchar(64);unique;not null"`
	Nickname string      `gorm:"types:varchar(128)"`
	Passwd   string      `gorm:"types:varchar(32);not null"`
	Skey     string      `gorm:"types:varchar(16);not null"`
	Headurl  string      `gorm:"types:varchar(128);unique;not null"`
	Uptime   int64       `gorm:"types:datetime"`
}

// TableName gorm use this to get tablename
// NOTE : it only works int where caulse
func (u User) TableName() string {
	var value int
	for _, c := range []rune(u.Username) {
		value = value + int(c)
	}
	return fmt.Sprintf("userinfo_tab_%d", value % 20)
}
