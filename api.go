package main

import (
	"bytes"
	"encoding/base64"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"messenger/models"
	"net/http"
	"time"
)

type Api struct {
	db *gorm.DB
	redis *redis.Client
}


func (api *Api) Init() {
	initRandom()
	if result := api.db.AutoMigrate(&models.User{}); result.Error != nil {
		panic(result.Error)
	}
}


func (api *Api) CreateSession(nickname string, r *http.Request) {

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


func (api *Api) Auth(nickname string, r *http.Request, w http.ResponseWriter) error {
	salt := RandStringRunes(20)
	api.redis.Set(nickname, salt, time.Second * 600)

	cookie := http.Cookie{
		Name: "sid",
		Value: base64.StdEncoding.EncodeToString(Hash(nickname + salt)),
	}
	http.SetCookie(w, &cookie)
	http.SetCookie(w, &http.Cookie{
		Name: "nickname",
		Value: nickname,
	})
	return nil
}


func (api *Api) CheckAuth(r *http.Request, w http.ResponseWriter) (string, error) {
	sidCookie, err := r.Cookie("sid")
	if err != nil {
		return "", nil
	}
	nicknameCookie, err := r.Cookie("nickname")
	if err != nil {
		return "", nil
	}
	salt := api.redis.Get(nicknameCookie.Value)
	con := nicknameCookie.Value + salt.Val()
	if base64.StdEncoding.EncodeToString(Hash(con)) == sidCookie.Value {
		return nicknameCookie.Value, nil
	}
	return "", nil
}
