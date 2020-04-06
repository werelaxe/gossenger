package backend

import (
	"encoding/json"
	"log"
	"messenger/common"
	"messenger/dbapi"
	"messenger/models"
	"net/http"
)

func SendMessageHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := EnsureLogin(api, writer, request)
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

		if err := api.SendMessage(sendMessageData.Text, user.ID, sendMessageData.ChatId); err != nil {
			log.Println("Can not send message: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}

func ListMessagesHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := EnsureLogin(api, writer, request)
		if user == nil {
			writer.WriteHeader(400)
			return
		}

		var listMessagesData models.ListMessagesRequestSchema
		if err := json.NewDecoder(request.Body).Decode(&listMessagesData); err != nil {
			log.Println("Can not list messages: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		ok, err := api.IsUserChatMember(user.ID, listMessagesData.ChatId)
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

		messages, err := api.ListMessages(listMessagesData.ChatId)
		if err != nil {
			log.Println("Can not list messages: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		var listMessagesResponseData models.ListMessagesResponseSchema
		for _, message := range messages {
			listMessagesResponseData = append(listMessagesResponseData, models.MessageResponseSchema{
				Text:        message.Text,
				SenderRefer: message.SenderRefer,
				Time:        message.Time,
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
