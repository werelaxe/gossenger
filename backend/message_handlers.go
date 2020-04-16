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
	"strconv"
	"time"
)

func processAllChatMembers(api *dbapi.Api, chat *models.Chat, callback func([]*models.User)) error {
	offset := 0
	for {
		users, err := api.ListChatMembers(chat, common.MaxApiLimit, offset)
		if err != nil {
			return errors.New("can not list chat members: " + err.Error())
		}
		if len(users) > 0 {
			callback(users)
		} else {
			break
		}
		offset += common.MaxApiLimit
	}
	return nil
}

func SendMessageHandler(api *dbapi.Api, connKeeper common.ConnectionKeeper) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := EnsureLogin(api, request)
		if loggedUser == nil {
			writer.WriteHeader(400)
			return
		}

		var sendMessageData models.SendMessageRequestSchema
		if err := json.NewDecoder(request.Body).Decode(&sendMessageData); err != nil {
			log.Println("Can not send message: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		escapedMessageText := html.EscapeString(sendMessageData.Text)

		chat, err := api.GetChat(sendMessageData.ChatId)
		if err != nil {
			log.Println("Can not send message: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if err := api.SendMessage(escapedMessageText, loggedUser.ID, sendMessageData.ChatId); err != nil {
			log.Println("Can not send message: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		fastMessageResponseData := models.FastMessageResponseSchema{
			Text:     escapedMessageText,
			SenderId: loggedUser.ID,
			Time:     time.Now().Unix(),
			ChatId:   sendMessageData.ChatId,
		}

		rawFastMessageResponseData, err := json.Marshal(fastMessageResponseData)
		if err != nil {
			log.Println("Can not marshal message response data after message sending")
			return
		}

		chatMembersCallback := func(users []*models.User) {
			for _, user := range users {
				conn, ok := connKeeper[common.MessagesConnType][user.ID]
				if !ok {
					log.Printf("Can not get connection for loggedUser with ID=%v\n", user.ID)
				} else {

					if err := conn.WriteMessage(1, rawFastMessageResponseData); err != nil {
						log.Println("Can not write to the loggedUser connection: " + err.Error())
					}
				}
			}
		}

		if err := processAllChatMembers(api, chat, chatMembersCallback); err != nil {
			log.Printf("Can not process all chat members, chat: %v, error: %v", chat, err)
			return
		}
	}
}

func ListMessagesHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := EnsureLogin(api, request)
		if loggedUser == nil {
			writer.WriteHeader(400)
			return
		}

		rawChatId, ok := request.URL.Query()["chat_id"]
		if !ok {
			log.Println("Can not list messages: there is no chat_id parameter")
			writer.WriteHeader(400)
			return
		}

		chatId, err := strconv.ParseUint(rawChatId[0], 10, 64)
		if err != nil {
			log.Println("Can not list messages: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		ok, err = api.IsUserChatMember(loggedUser.ID, uint(chatId))
		if err != nil {
			log.Println("Can not list messages: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if !ok {
			log.Println("Can not list messages: loggedUser is not a member of requested chat")
			writer.WriteHeader(400)
			return
		}

		limit, offset, err := common.GetLimitAndOffset(request.URL.Query())
		if err != nil {
			log.Println("Can not list messages: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		messages, err := api.ListMessages(uint(chatId), limit, offset)
		if err != nil {
			log.Println("Can not list messages: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		listMessagesResponseData := models.ListMessagesResponseSchema{}
		for _, message := range messages {
			listMessagesResponseData = append(listMessagesResponseData, models.MessageResponseSchema{
				Text:     message.Text,
				SenderId: message.SenderRefer,
				Time:     message.Time,
			})
		}

		rawListMessagesResponseData, err := json.Marshal(listMessagesResponseData)
		if err != nil {
			log.Println("Can not list messages: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		_, err = writer.Write(rawListMessagesResponseData)
		if err != nil {
			log.Println("Can not list messages: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
