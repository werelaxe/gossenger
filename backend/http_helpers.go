package backend

import (
	"encoding/base64"
	"log"
	"messenger/common"
	"messenger/dbapi"
	"messenger/models"
	"net/http"
)

func Redirect(writer http.ResponseWriter, path string) {
	if _, err := writer.Write([]byte("<script>" + path + "</script>")); err != nil {
		log.Println("Can not redirect: " + err.Error())
	}
}

func Auth(nickname string, w http.ResponseWriter, salt string) error {
	cookie := http.Cookie{
		Name:  "sid",
		Value: base64.StdEncoding.EncodeToString(common.Hash(nickname + salt)),
	}
	http.SetCookie(w, &cookie)
	http.SetCookie(w, &http.Cookie{
		Name:  "nickname",
		Value: nickname,
	})
	return nil
}

func CheckAuth(api *dbapi.Api, r *http.Request) (string, error) {
	sidCookie, err := r.Cookie("sid")
	if err != nil {
		return "", nil
	}
	nicknameCookie, err := r.Cookie("nickname")
	if err != nil {
		return "", nil
	}
	salt := api.Redis.Get(nicknameCookie.Value)
	con := nicknameCookie.Value + salt.Val()
	if base64.StdEncoding.EncodeToString(common.Hash(con)) == sidCookie.Value {
		return nicknameCookie.Value, nil
	}
	return "", nil
}

func EnsureLogin(api *dbapi.Api, writer http.ResponseWriter, request *http.Request) *models.User {
	nickname, err := CheckAuth(api, request)
	if err != nil {
		log.Println("Can not index: " + err.Error())
		writer.WriteHeader(400)
		return nil
	}
	if nickname == "" {
		log.Println("Wrong cookie")
		writer.WriteHeader(400)
		return nil
	}
	user, err := api.GetUserByNickname(nickname)
	if err != nil {
		log.Println("Can not get user: " + err.Error())
		writer.WriteHeader(400)
		return nil
	}
	return user
}
