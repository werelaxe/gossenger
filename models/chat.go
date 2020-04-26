package models

import "github.com/jinzhu/gorm"

type Chat struct {
	gorm.Model
	Title      string
	AdminRefer uint
	Messages   []Message `gorm:"foreignkey:ChatRefer"`
	Members    []*User   `gorm:"many2many:chat_members;"`
	IsPrivate  bool
}

type PrivateRelation struct {
	gorm.Model
	ChatRefer       uint
	FirstUserRefer  uint `gorm:"unique_index:users_pair"`
	SecondUserRefer uint `gorm:"unique_index:users_pair"`
}
