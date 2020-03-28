package main

import (
	"encoding/json"
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

		salt := api.CreateSession(registerUserData.Nickname)
		if err := Auth(registerUserData.Nickname, writer, salt); err != nil {
			log.Println("Can not auth after registration: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}

func indexHandler(api *Api) HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := ensureLogin(api, writer, request)
		if user == nil {
			return
		}

		if _, err := writer.Write([]byte("Hello, " + user.FirstName + " " + user.LastName)); err != nil {
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
		if err := json.NewDecoder(request.Body).Decode(&loginUserData); err != nil {
			if e, ok := err.(*json.SyntaxError); ok {
				log.Printf("syntax error at byte offset %d", e.Offset)
			}
			log.Println("Can not login user: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		ok, err := api.IsValidPair(loginUserData.Nickname, loginUserData.Password)
		if err != nil {
			log.Println("Can not check pair: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if !ok {
			log.Println("Invalid login/password")
			writer.WriteHeader(400)
			return
		}

		salt := api.CreateSession(loginUserData.Nickname)

		if err := Auth(loginUserData.Nickname, writer, salt); err != nil {
			log.Println("Can not login user: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}

func createChatHandler(api *Api) HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := ensureLogin(api, writer, request)
		if user == nil {
			return
		}

		var createChatData models.CreateChatSchema
		if err := json.NewDecoder(request.Body).Decode(&createChatData); err != nil {
			log.Println("Can not create chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		var users []*models.User

		for memberId, _ := range unique(createChatData.Members) {
			member, err := api.GetUserById(memberId)
			if err != nil {
				log.Println("Can not create chat: " + err.Error())
				writer.WriteHeader(400)
				return
			}
			users = append(users, member)
		}

		if err := api.CreateChat(createChatData.Title, user, users); err != nil {
			log.Println("Can not create chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}

func addUserToChatHandler(api *Api) HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := ensureLogin(api, writer, request)
		if user == nil {
			return
		}

		var addUserToChatData models.AddUserToChatSchema
		if err := json.NewDecoder(request.Body).Decode(&addUserToChatData); err != nil {
			log.Println("Can add user to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		newMember, err := api.GetUserById(addUserToChatData.UserId)
		if err != nil {
			log.Println("Can add user to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		chat, err := api.GetChat(addUserToChatData.ChatId)
		if err != nil {
			log.Println("Can add user to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		users, err := api.ListChatMembers(chat)
		if err != nil {
			log.Println("Can add user to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		uniqueUserIds := getUniqueUserIds(users)

		if _, ok := uniqueUserIds[user.ID]; !ok {
			log.Println("Can add user to chat: logged user must be in chat")
			writer.WriteHeader(400)
			return
		}

		if _, ok := uniqueUserIds[newMember.ID]; ok {
			log.Println("Can add user to chat: user is already in chat")
			writer.WriteHeader(400)
			return
		}

		if err = api.AddUserToChat(newMember, chat); err != nil {
			log.Println("Can add user to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
