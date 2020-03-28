package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Nickname     string `gorm:"unique_index"`
	FirstName    string
	LastName     string
	PasswordHash []byte
	SentMessages []Message `gorm:"foreignkey:SenderRefer"`
	Chats        []*Chat   `gorm:"many2many:chat_members;"`
}
