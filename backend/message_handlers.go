package backend

import (
	"encoding/json"
	"gossenger/common"
	"gossenger/dbapi"
	"gossenger/models"
	"log"
	"net/http"
	"strconv"
	"time"
)

func SendMessageHandler(api *dbapi.Api, reminder common.UpgradeReminder) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := EnsureLogin(api, request)
		if user == nil {
			writer.WriteHeader(400)
			return
		}

		var sendMessageData models.SendMessageRequestSchema
		if err := json.NewDecoder(request.Body).Decode(&sendMessageData); err != nil {
			log.Println("Can not send message: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		chat, err := api.GetChat(sendMessageData.ChatId)
		if err != nil {
			log.Println("Can not send message: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if err := api.SendMessage(sendMessageData.Text, user.ID, sendMessageData.ChatId); err != nil {
			log.Println("Can not send message: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		users, err := api.ListChatMembers(chat)
		if err != nil {
			log.Println("Can not list chat members after message sending")
			return
		}

		fastMessageResponseData := models.FastMessageResponseSchema{
			Text:     sendMessageData.Text,
			SenderId: user.ID,
			Time:     time.Now().Unix(),
			ChatId:   sendMessageData.ChatId,
		}

		rawFastMessageResponseData, err := json.Marshal(fastMessageResponseData)
		if err != nil {
			log.Println("Can not marshal message response data after message sending")
			return
		}

		for _, user := range users {
			conn, ok := reminder[user.ID]
			if !ok {
				log.Printf("Can not get connection for user with ID=%v\n", user.ID)
			} else {

				if err := conn.WriteMessage(1, rawFastMessageResponseData); err != nil {
					log.Println("Can not write to the user connection: " + err.Error())
				}
			}
		}
	}
}

func ListMessagesHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := EnsureLogin(api, request)
		if user == nil {
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

		ok, err = api.IsUserChatMember(user.ID, uint(chatId))
		if err != nil {
			log.Println("Can not list messages: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if !ok {
			log.Println("Can not list messages: user is not a member of requested chat")
			writer.WriteHeader(400)
			return
		}

		messages, err := api.ListMessages(uint(chatId))
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
