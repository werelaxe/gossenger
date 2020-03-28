package main

import (
	"bytes"
	"encoding/base64"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"messenger/models"
	"time"
)

type Api struct {
	db    *gorm.DB
	redis *redis.Client
}

func (api *Api) Init() {
	initRandom()
	if result := api.db.AutoMigrate(&models.User{}); result.Error != nil {
		panic(result.Error)
	}
}

func (api *Api) RegisterUser(nickname, firstName, lastName, password string) error {
	result := api.db.Create(&models.User{
		Nickname:     nickname,
		FirstName:    firstName,
		LastName:     lastName,
		PasswordHash: Hash(password),
	})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (api *Api) IsValidPair(nickname, password string) (bool, error) {
	var user models.User

	if err := api.db.Where("nickname = ?", nickname).First(&user).Error; err != nil {
		return false, err
	}
	return bytes.Equal(user.PasswordHash, Hash(password)), nil
}

func (api *Api) CreateSession(nickname string) string {
	salt := RandStringRunes(20)
	api.redis.Set(nickname, salt, time.Second*600)
	return salt
}

func (api *Api) ValidateSession(nickname, sid string) bool {
	salt := api.redis.Get(nickname)
	con := nickname + salt.Val()
	if base64.StdEncoding.EncodeToString(Hash(con)) == sid {
		return false
	}
	return true
}
