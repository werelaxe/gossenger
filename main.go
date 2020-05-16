package main

import (
	"github.com/gorilla/websocket"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gossenger/backend"
	"gossenger/common"
	"gossenger/dbapi"
	"gossenger/frontend"
	"log"
	"net/http"
	"os"
)

const defaultConfigName = "dev.config"

func main() {
	configName := defaultConfigName
	if len(os.Args) >= 2 {
		configName = os.Args[1]
	}
	config, err := common.GetConfig(configName)
	if err != nil {
		panic(err)
	}

	api, err := dbapi.GetApiByConfig(config)
	if err != nil {
		panic(err)
	}
	defer api.Close()

	upgrader := websocket.Upgrader{}
	connectionKeeper := common.ConnectionKeeper{}
	connectionKeeper.Init()
	defer connectionKeeper.Close()

	var templateManager frontend.TemplateManager
	templateManager.Init("frontend/templates")

	fs := http.FileServer(http.Dir("./frontend/static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/register", backend.RegisterHandler(api))

	http.HandleFunc("/login", backend.LoginHandler(api))

	http.HandleFunc("/chats/create", backend.CreateChatHandler(api, connectionKeeper))
	http.HandleFunc("/chats/create_private", backend.CreatePrivateChatHandler(api, connectionKeeper))
	http.HandleFunc("/chats/add_user", backend.AddUserToChatHandler(api))
	http.HandleFunc("/chats/list", backend.ListChatsHandler(api))
	http.HandleFunc("/chats/show", backend.ShowChatHandler(api))
	http.HandleFunc("/chats/list_members", backend.ListChatMembersHandler(api))

	http.HandleFunc("/messages/send", backend.SendMessageHandler(api, connectionKeeper))
	http.HandleFunc("/messages/list", backend.ListMessagesHandler(api))

	http.HandleFunc("/users/list", backend.ListUsersHandler(api))
	http.HandleFunc("/users/search", backend.SearchUsersHandler(api))
	http.HandleFunc("/users/show", backend.ShowUserHandler(api))

	http.HandleFunc("/login_page", frontend.LoginPageHandler(api, &templateManager))
	http.HandleFunc("/register_page", frontend.RegisterPageHandler(api, &templateManager))
	http.HandleFunc("/users_page", frontend.UsersPageHandler(api, &templateManager))
	http.HandleFunc("/user_page", frontend.UserPageHandler(api, &templateManager))
	http.HandleFunc("/", frontend.IndexHandler(api, &templateManager))

	http.HandleFunc("/messages_ws", backend.WebSocketHandler(api, &upgrader, connectionKeeper, common.MessagesConnType))
	http.HandleFunc("/chats_ws", backend.WebSocketHandler(api, &upgrader, connectionKeeper, common.ChatsConnType))

	log.Fatal(http.ListenAndServe(common.GetAddr(config.Server.Host, config.Server.Port), nil))
}
