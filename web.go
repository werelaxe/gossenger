package main

import (
	"encoding/base64"
	"log"
	"net/http"
)

func redirect(writer http.ResponseWriter, request *http.Request, path string) {
	if _, err := writer.Write([]byte("<script>" + path + "</script>")); err != nil {
		log.Println("Can not redirect: " + err.Error())
	}
}

func Auth(nickname string, r *http.Request, w http.ResponseWriter, salt string) error {
	cookie := http.Cookie{
		Name:  "sid",
		Value: base64.StdEncoding.EncodeToString(Hash(nickname + salt)),
	}
	http.SetCookie(w, &cookie)
	http.SetCookie(w, &http.Cookie{
		Name:  "nickname",
		Value: nickname,
	})
	return nil
}

func CheckAuth(api *Api, r *http.Request, w http.ResponseWriter) (string, error) {
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

func ensureLogin(api *Api, writer http.ResponseWriter, request *http.Request) (string, bool) {
	nickname, err := CheckAuth(api, request, writer)
	if err != nil {
		log.Println("Can not index: " + err.Error())
		writer.WriteHeader(400)
		return "", false
	}
	if nickname == "" {
		log.Println("Wrong cookie")
		writer.WriteHeader(400)
		return "", false
	}
	return nickname, true
}
