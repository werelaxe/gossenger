package backend

import (
	"github.com/gorilla/websocket"
	"gossenger/common"
	"gossenger/dbapi"
	"log"
	"net/http"
)

func MessagesHandler(api *dbapi.Api, upgrader *websocket.Upgrader, um common.UpgradeReminder) common.HandlerFuncType {
	return func(w http.ResponseWriter, r *http.Request) {
		user := EnsureLogin(api, r)
		if user == nil {
			return
		}

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Can not get connection: ", err.Error())
			return
		}

		log.Printf("New connection created for user with ID=%v\n", user.ID)
		um[user.ID] = c
	}
}
