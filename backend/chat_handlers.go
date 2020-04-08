package backend

import (
	"encoding/json"
	"gossenger/common"
	"gossenger/dbapi"
	"gossenger/models"
	"log"
	"net/http"
	"strconv"
)

func CreateChatHandler(api *dbapi.Api, connKeeper common.ConnectionKeeper) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := EnsureLogin(api, request)
		if user == nil {
			writer.WriteHeader(400)
			return
		}

		var createChatData models.CreateChatRequestSchema
		if err := json.NewDecoder(request.Body).Decode(&createChatData); err != nil {
			log.Println("Can not create chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		var users []*models.User

		for memberId := range common.Unique(createChatData.Members) {
			member, err := api.GetUserById(memberId)
			if err != nil {
				log.Println("Can not create chat: " + err.Error())
				writer.WriteHeader(400)
				return
			}
			users = append(users, member)
		}

		newChatId, err := api.CreateChat(createChatData.Title, user, users)
		if err != nil {
			log.Println("Can not create chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		fastChatCreatingResponseData := models.FastChatCreatingResponseSchema{
			Title: createChatData.Title,
			ID:    newChatId,
		}

		rawFastChatCreatingResponseData, err := json.Marshal(fastChatCreatingResponseData)
		if err != nil {
			log.Println("Can not marshal chat creating response data after message sending")
			return
		}

		for _, userId := range createChatData.Members {
			conn, ok := connKeeper[common.ChatsConnType][userId]
			if !ok {
				log.Printf("Can not get connection for user with ID=%v\n", user.ID)
			} else {
				if err := conn.WriteMessage(1, rawFastChatCreatingResponseData); err != nil {
					log.Println("Can not write to the user connection: " + err.Error())
				}
			}
		}
	}
}

func AddUserToChatHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := EnsureLogin(api, request)
		if user == nil {
			writer.WriteHeader(400)
			return
		}

		var addUserToChatData models.AddUserToChatRequestSchema
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

func ListUserChatsHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := EnsureLogin(api, request)
		if user == nil {
			writer.WriteHeader(400)
			return
		}

		chats, err := api.ListUserChats(user)
		if err != nil {
			log.Println("Can not list user chats: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		var listChatsResponseData models.ListChatsResponseSchema
		for _, chat := range chats {
			listChatsResponseData = append(listChatsResponseData, models.ChatResponseSchema{
				ChatId: chat.ID,
				Title:  chat.Title,
			})
		}

		rawListChatsResponseData, err := json.Marshal(listChatsResponseData)
		if err != nil {
			log.Println("Can not marshal listChatsResponseData: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		_, err = writer.Write(rawListChatsResponseData)
		if err != nil {
			log.Println("Can not write listChatsResponseData: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}

func ListChatMembersHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := EnsureLogin(api, request)
		if user == nil {
			writer.WriteHeader(400)
			return
		}

		rawChatId, ok := request.URL.Query()["chat_id"]

		if !ok {
			log.Println("Can not list chat members: there is no chat_id parameter")
			writer.WriteHeader(400)
			return
		}

		chatId, err := strconv.ParseUint(rawChatId[0], 10, 64)
		if err != nil {
			log.Println("Can not list chat members: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		ok, err = api.IsUserChatMember(user.ID, uint(chatId))
		if err != nil {
			log.Println("Can not list chat members: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if !ok {
			log.Println("Can not list chat members: user is not a member of requested chat")
			writer.WriteHeader(400)
			return
		}

		chat, err := api.GetChat(uint(chatId))
		if err != nil {
			log.Println("Can not list chat members: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		members, err := api.ListChatMembers(chat)
		if err != nil {
			log.Println("Can not list chat members: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		var listChatMembersResponseData models.ListChatMembersResponseSchema
		for _, member := range members {
			listChatMembersResponseData = append(listChatMembersResponseData, models.ChatMemberResponseSchema{
				ID:        member.ID,
				Nickname:  member.Nickname,
				FirstName: member.FirstName,
				LastName:  member.LastName,
			})
		}

		rawListChatMembersResponseData, err := json.Marshal(listChatMembersResponseData)
		if err != nil {
			log.Println("Can not list chat members: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		_, err = writer.Write(rawListChatMembersResponseData)
		if err != nil {
			log.Println("Can not list chat members: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
