package models

import "github.com/jinzhu/gorm"

type Message struct {
	gorm.Model
	Text        string
	SenderRefer int64
	ChatRefer   int64
	Time        int64
}
