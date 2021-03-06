package backend

import (
	"encoding/json"
	"errors"
	"gossenger/common"
	"gossenger/dbapi"
	"gossenger/models"
	"html"
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

		createChatData.Title = html.EscapeString(createChatData.Title)
		createChatData.Members = append(createChatData.Members, loggedUser.ID)
		newChatId, err := api.CreateChat(&createChatData, loggedUser)
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

		for _, userId := range append(createChatData.Members) {
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

		users, err := api.ListChatMembers(chat, common.MaxApiLimit, 0)
		if err != nil {
			log.Println("Can not add loggedUser to chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		uniqueUserIds, err := dbapi.GetUniqueUserIds(users)
		if err != nil {
			log.Println("Can not get unique user ids: " + err.Error())
			writer.WriteHeader(400)
			return
		}

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

		limit, offset, err := common.GetLimitAndOffset(request.URL.Query())
		if err != nil {
			log.Println("Can not list chats: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		chats, err := api.ListChats(loggedUser, limit, offset)
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

			chatTitle := "Chat title"
			if chat.IsPrivate {
				anotherUser, err := api.GetAnotherUserFromPrivateChat(&chat, loggedUser.ID)
				if err != nil {
					log.Println("Can not get another user for private chat display name")
				} else {
					chatTitle = anotherUser.FirstName + " " + anotherUser.LastName
				}
			} else {
				chatTitle = chat.Title
			}
			if lastMessage != nil {
				previewMessageText := lastMessage.Text
				if len(previewMessageText) > 40 {
					previewMessageText = previewMessageText[:37] + "..."
				}

				listChatsResponseData = append(listChatsResponseData, models.ChatResponseSchema{
					ChatId:               chat.ID,
					Title:                chatTitle,
					PreviewMessageText:   previewMessageText,
					PreviewMessageSender: lastMessage.SenderRefer,
				})
				sortingMap[chat.ID] = TimePair{lastMessage.Time, 0}
			} else {
				listChatsResponseData = append(listChatsResponseData, models.ChatResponseSchema{
					ChatId:               chat.ID,
					Title:                chatTitle,
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
				return firstElement.ChatCreationTime < secondElement.ChatCreationTime
			} else if firstElement.LastMessageTime != 0 {
				return firstElement.LastMessageTime < secondElement.ChatCreationTime
			} else {
				return firstElement.ChatCreationTime < secondElement.LastMessageTime
			}
		})

		if len(listChatsResponseData) > limit {
			listChatsResponseData = listChatsResponseData[:limit]
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

		members, err := api.ListChatMembers(chat, common.MaxApiLimit, 0)
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

func getChatResponseData(api *dbapi.Api, chat *models.Chat) (*models.ChatResponseSchema, error) {
	lastMessage, err := api.GetChatLastMessage(chat.ID)
	if err != nil {
		return nil, errors.New("can not get chat response data: " + err.Error())
	}
	if lastMessage != nil {
		previewMessageText := lastMessage.Text
		if len(previewMessageText) > 40 {
			previewMessageText = previewMessageText[:37] + "..."
		}
		return &models.ChatResponseSchema{
			ChatId:               chat.ID,
			Title:                chat.Title,
			PreviewMessageText:   previewMessageText,
			PreviewMessageSender: lastMessage.SenderRefer,
		}, nil
	} else {
		return &models.ChatResponseSchema{
			ChatId:               chat.ID,
			Title:                chat.Title,
			PreviewMessageSender: chat.AdminRefer,
		}, nil
	}
}

func ShowChatHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := EnsureLogin(api, request)
		if loggedUser == nil {
			writer.WriteHeader(400)
			return
		}

		rawChatId, ok := request.URL.Query()["chat_id"]

		if !ok {
			log.Println("Can not show chat: there is no chat_id parameter")
			writer.WriteHeader(400)
			return
		}

		chatId, err := strconv.ParseUint(rawChatId[0], 10, 64)
		if err != nil {
			log.Println("Can not show chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		ok, err = api.IsUserChatMember(loggedUser.ID, uint(chatId))
		if err != nil {
			log.Println("Can not show chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if !ok {
			log.Println("Can not show chat: loggedUser is not a member of requested chat")
			writer.WriteHeader(400)
			return
		}

		chat, err := api.GetChat(uint(chatId))
		if err != nil {
			log.Println("Can not show chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		chatResponseData, err := getChatResponseData(api, chat)

		rawChatResponseData, err := json.Marshal(chatResponseData)
		if err != nil {
			log.Println("Can not show chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		_, err = writer.Write(rawChatResponseData)
		if err != nil {
			log.Println("Can not show chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}

func CreatePrivateChatHandler(api *dbapi.Api, connKeeper common.ConnectionKeeper) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := EnsureLogin(api, request)
		if loggedUser == nil {
			writer.WriteHeader(400)
			return
		}

		var createPrivateChatData models.CreatePrivateChatRequestSchema
		if err := json.NewDecoder(request.Body).Decode(&createPrivateChatData); err != nil {
			log.Println("Can not create private chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		privateRelation, err := api.GetPrivateRelation(createPrivateChatData.UserId, loggedUser.ID)
		if err != nil {
			log.Println("Can not create private chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if privateRelation != nil {
			log.Println("Can not create private chat: private relation is already exists")
			writer.WriteHeader(400)
			writer.Write([]byte("is already exists: " + strconv.Itoa(int(privateRelation.ChatRefer))))
			return
		}

		chatId, err := api.CreateChat(&models.CreateChatRequestSchema{
			Title:     "private chat",
			Members:   []uint{loggedUser.ID, createPrivateChatData.UserId},
			IsPrivate: true,
		}, loggedUser)

		if err != nil {
			log.Println("Can not create private chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		rawResponseData, err := json.Marshal(struct {
			ChatId uint `json:"chat_id"`
		}{chatId})
		if err != nil {
			log.Println("Can not send response after creating private chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if _, err = writer.Write(rawResponseData); err != nil {
			log.Println("Can not send response after creating private chat: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		anotherUser, err := api.GetUserById(createPrivateChatData.UserId)
		if err != nil {
			log.Println("Can not get user by id for private chat title")
			return
		}

		loggedUserFastChatCreatingResponseData := models.FastChatCreatingResponseSchema{
			Title:                anotherUser.FirstName + " " + anotherUser.LastName,
			ID:                   chatId,
			PreviewMessageSender: loggedUser.ID,
			PreviewMessageText:   "",
		}

		rawFastLoggedUserChatCreatingResponseData, err := json.Marshal(loggedUserFastChatCreatingResponseData)
		if err != nil {
			log.Println("Can not marshal private chat creating response data after message sending for logged user")
			return
		}

		loggedUserConn, ok := connKeeper[common.ChatsConnType][loggedUser.ID]
		if !ok {
			log.Printf("Can not get connection for loggedUser with ID=%v\n", loggedUser.ID)
		} else {
			if err := loggedUserConn.WriteMessage(1, rawFastLoggedUserChatCreatingResponseData); err != nil {
				log.Println("Can not write to the loggedUser connection: " + err.Error())
			}
		}

		anotherUserFastChatCreatingResponseData := models.FastChatCreatingResponseSchema{
			Title:                loggedUser.FirstName + " " + loggedUser.LastName,
			ID:                   chatId,
			PreviewMessageSender: loggedUser.ID,
			PreviewMessageText:   "",
		}

		rawFastAnotherUserChatCreatingResponseData, err := json.Marshal(anotherUserFastChatCreatingResponseData)
		if err != nil {
			log.Println("Can not marshal private chat creating response data after message sending for another user")
			return
		}
		anotherUserConn, ok := connKeeper[common.ChatsConnType][anotherUser.ID]
		if !ok {
			log.Printf("Can not get connection for another user with ID=%v\n", loggedUser.ID)
		} else {
			if err := anotherUserConn.WriteMessage(1, rawFastAnotherUserChatCreatingResponseData); err != nil {
				log.Println("Can not write to the another user connection: " + err.Error())
			}
		}
	}
}
