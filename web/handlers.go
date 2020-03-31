package web

import (
	"encoding/json"
	"log"
	"messenger/dbapi"
	"messenger/models"
	"messenger/utils"
	"net/http"
)

type HandlerFuncType func(writer http.ResponseWriter, request *http.Request)

func RegisterHandler(api *dbapi.Api) HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			log.Println("Wrong method (should be POST)")
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

func IndexHandler(api *dbapi.Api) HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := EnsureLogin(api, writer, request)
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

func LoginHandler(api *dbapi.Api) HandlerFuncType {
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

func CreateChatHandler(api *dbapi.Api) HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := EnsureLogin(api, writer, request)
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

		for memberId, _ := range utils.Unique(createChatData.Members) {
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

func AddUserToChatHandler(api *dbapi.Api) HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := EnsureLogin(api, writer, request)
		if user == nil {
			return
		}

		var addUserToChatData models.AddUserToChatSchema
		if err := json.NewDecoder(request.Body).Decode(&addUserToChatData); err != nil {
			log.Println("Can not add user to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		newMember, err := api.GetUserById(addUserToChatData.UserId)
		if err != nil {
			log.Println("Can not add user to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		chat, err := api.GetChat(addUserToChatData.ChatId)
		if err != nil {
			log.Println("Can not add user to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		users, err := api.ListChatMembers(chat)
		if err != nil {
			log.Println("Can not add user to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		uniqueUserIds := dbapi.GetUniqueUserIds(users)

		if _, ok := uniqueUserIds[user.ID]; !ok {
			log.Println("Can not add user to chat: logged user must be in chat")
			writer.WriteHeader(400)
			return
		}

		if _, ok := uniqueUserIds[newMember.ID]; ok {
			log.Println("Can not add user to chat: user is already in chat")
			writer.WriteHeader(400)
			return
		}

		if err = api.AddUserToChat(newMember, chat); err != nil {
			log.Println("Can not add user to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}

func SendMessageHandler(api *dbapi.Api) HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := EnsureLogin(api, writer, request)
		if user == nil {
			return
		}

		var sendMessageData models.SendMessageSchema
		if err := json.NewDecoder(request.Body).Decode(&sendMessageData); err != nil {
			log.Println("Can not send message: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if err := api.SendMessage(sendMessageData.Text, user.ID, sendMessageData.ChatId); err != nil {
			log.Println("Can not send message: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
