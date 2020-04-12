package backend

import (
	"encoding/json"
	"gossenger/common"
	"gossenger/dbapi"
	"gossenger/models"
	"log"
	"net/http"
	"sort"
	"strconv"
)

func CreateChatHandler(api *dbapi.Api, connKeeper common.ConnectionKeeper) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := EnsureLogin(api, request)
		if loggedUser == nil {
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

		for memberId := range common.Unique(append(createChatData.Members, loggedUser.ID)) {
			member, err := api.GetUserById(memberId)
			if err != nil {
				log.Println("Can not create chat: " + err.Error())
				writer.WriteHeader(400)
				return
			}
			users = append(users, member)
		}

		newChatId, err := api.CreateChat(createChatData.Title, loggedUser, users)
		if err != nil {
			log.Println("Can not create chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		rawResponseData, err := json.Marshal(struct {
			ChatId uint `json:"chat_id"`
		}{newChatId})
		if err != nil {
			log.Println("Can not send response after created chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if _, err = writer.Write(rawResponseData); err != nil {
			log.Println("Can not send response after created chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		fastChatCreatingResponseData := models.FastChatCreatingResponseSchema{
			Title:                createChatData.Title,
			ID:                   newChatId,
			PreviewMessageSender: loggedUser.ID,
		}

		rawFastChatCreatingResponseData, err := json.Marshal(fastChatCreatingResponseData)
		if err != nil {
			log.Println("Can not marshal chat creating response data after message sending")
			return
		}

		for _, userId := range append(createChatData.Members, loggedUser.ID) {
			conn, ok := connKeeper[common.ChatsConnType][userId]
			if !ok {
				log.Printf("Can not get connection for loggedUser with ID=%v\n", userId)
			} else {
				if err := conn.WriteMessage(1, rawFastChatCreatingResponseData); err != nil {
					log.Println("Can not write to the loggedUser connection: " + err.Error())
				}
			}
		}
	}
}

func AddUserToChatHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := EnsureLogin(api, request)
		if loggedUser == nil {
			writer.WriteHeader(400)
			return
		}

		var addUserToChatData models.AddUserToChatRequestSchema
		if err := json.NewDecoder(request.Body).Decode(&addUserToChatData); err != nil {
			log.Println("Can not add loggedUser to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		newMember, err := api.GetUserById(addUserToChatData.UserId)
		if err != nil {
			log.Println("Can not add loggedUser to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		chat, err := api.GetChat(addUserToChatData.ChatId)
		if err != nil {
			log.Println("Can not add loggedUser to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		users, err := api.ListChatMembers(chat)
		if err != nil {
			log.Println("Can not add loggedUser to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		uniqueUserIds := dbapi.GetUniqueUserIds(users)

		if _, ok := uniqueUserIds[loggedUser.ID]; !ok {
			log.Println("Can not add loggedUser to chat: logged loggedUser must be in chat")
			writer.WriteHeader(400)
			return
		}

		if _, ok := uniqueUserIds[newMember.ID]; ok {
			log.Println("Can not add loggedUser to chat: loggedUser is already in chat")
			writer.WriteHeader(400)
			return
		}

		if err = api.AddUserToChat(newMember, chat); err != nil {
			log.Println("Can not add loggedUser to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}

func ListChatsHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := EnsureLogin(api, request)
		if loggedUser == nil {
			writer.WriteHeader(400)
			return
		}

		chats, err := api.ListChats(loggedUser)
		if err != nil {
			log.Println("Can not list chats: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		type TimePair struct {
			LastMessageTime  int64
			ChatCreationTime int64
		}

		sortingMap := make(map[uint]TimePair)

		var listChatsResponseData models.ListChatsResponseSchema
		for _, chat := range chats {
			lastMessage, err := api.GetChatLastMessage(chat.ID)
			if err != nil {
				log.Println("Can not list chats: " + err.Error())
				writer.WriteHeader(400)
				return
			}

			if lastMessage != nil {
				listChatsResponseData = append(listChatsResponseData, models.ChatResponseSchema{
					ChatId:               chat.ID,
					Title:                chat.Title,
					PreviewMessageText:   lastMessage.Text,
					PreviewMessageSender: lastMessage.SenderRefer,
				})
				sortingMap[chat.ID] = TimePair{lastMessage.Time, 0}
			} else {
				listChatsResponseData = append(listChatsResponseData, models.ChatResponseSchema{
					ChatId:               chat.ID,
					Title:                chat.Title,
					PreviewMessageSender: chat.AdminRefer,
				})
				sortingMap[chat.ID] = TimePair{0, chat.CreatedAt.Unix()}
			}
		}

		sort.Slice(listChatsResponseData, func(i, j int) bool {
			firstElement := sortingMap[listChatsResponseData[j].ChatId]
			secondElement := sortingMap[listChatsResponseData[i].ChatId]
			if firstElement.LastMessageTime != 0 && secondElement.LastMessageTime != 0 {
				return firstElement.LastMessageTime < secondElement.LastMessageTime
			} else if firstElement.LastMessageTime == 0 && secondElement.LastMessageTime == 0 {
				return firstElement.ChatCreationTime > secondElement.ChatCreationTime
			} else if firstElement.LastMessageTime != 0 {
				return firstElement.LastMessageTime < secondElement.ChatCreationTime
			} else {
				return firstElement.ChatCreationTime < secondElement.LastMessageTime
			}
		})

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
		loggedUser := EnsureLogin(api, request)
		if loggedUser == nil {
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

		ok, err = api.IsUserChatMember(loggedUser.ID, uint(chatId))
		if err != nil {
			log.Println("Can not list chat members: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if !ok {
			log.Println("Can not list chat members: loggedUser is not a member of requested chat")
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
