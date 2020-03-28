package models

import "github.com/jinzhu/gorm"

type Message struct {
	gorm.Model
	Text        string
	SenderRefer uint
	ChatRefer   uint
	Time        int64
}
