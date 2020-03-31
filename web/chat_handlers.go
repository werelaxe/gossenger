package web

import (
	"encoding/json"
	"log"
	"messenger/dbapi"
	"messenger/models"
	"messenger/utils"
	"net/http"
)

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
