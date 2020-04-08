package backend

import (
	"github.com/gorilla/websocket"
	"gossenger/common"
	"gossenger/dbapi"
	"log"
	"net/http"
)

func WebSocketHandler(api *dbapi.Api, upgrader *websocket.Upgrader, connKeeper common.ConnectionKeeper, connType string) common.HandlerFuncType {
	return func(w http.ResponseWriter, r *http.Request) {
		user := EnsureLogin(api, r)
		if user == nil {
			return
		}

		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Can not get connection for '%v' handler: %v\n", connType, err.Error())
			return
		}

		log.Printf("New connection (type: %v) created for user with ID=%v\n", connType, user.ID)
		connKeeper[connType][user.ID] = c
	}
}
