package web

import (
	"encoding/json"
	"log"
	"messenger/dbapi"
	"messenger/models"
	"net/http"
)

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
