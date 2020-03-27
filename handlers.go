package main

import (
	"encoding/json"
	"fmt"
	"log"
	"messenger/models"
	"net/http"
)


type HandlerFuncType func(writer http.ResponseWriter, request *http.Request)


func registerHandler(api *Api) HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(400)
			return
		}

		var registerUserData models.RegisterUserSchema
		decoder := json.NewDecoder(request.Body)

		if err := decoder.Decode(&registerUserData); err != nil {
			if e, ok := err.(*json.SyntaxError); ok {
				log.Printf("syntax error at byte offset %d", e.Offset)
			}
			log.Println("Can not register user: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if err := api.RegisterUser(registerUserData.Nickname, registerUserData.FirstName, registerUserData.LastName, registerUserData.Password); err != nil {
			log.Println("Can not register user: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if err := api.Auth(registerUserData.Nickname, request, writer); err != nil {
			log.Println("Can not auth after registration: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}

func indexHandler(api *Api) HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		nickname, err := api.CheckAuth(request, writer)
		if err != nil {
			log.Println("Can not index: " + err.Error())
			writer.WriteHeader(400)
			return
		}
		if nickname == "" {
			if _, err := writer.Write([]byte("You are anon. Redirecting...")); err != nil {
				log.Println("Can not redirect: " + err.Error())
				writer.WriteHeader(400)
				return
			}
			log.Println("Wrong cookie")
			writer.WriteHeader(400)
			return
		}
		if _, err = writer.Write([]byte("Hello, " + nickname)); err != nil {
			log.Println("Can not write page: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}


func loginHandler(api *Api) HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(400)
			return
		}

		var loginUserData models.LoginUserSchema
		decoder := json.NewDecoder(request.Body)

		if err := decoder.Decode(&loginUserData); err != nil {
			if e, ok := err.(*json.SyntaxError); ok {
				log.Printf("syntax error at byte offset %d", e.Offset)
			}
			log.Println("Can not login user: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if err := api.Auth(loginUserData.Nickname, request, writer); err != nil {
			log.Println("Can not login user: " + err.Error())
			writer.WriteHeader(400)
			return
		}
		fmt.Println(loginUserData)
		fmt.Println("Logged in!")
	}
}
