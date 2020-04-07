package main

import (
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"gossenger/backend"
	"gossenger/common"
	"gossenger/dbapi"
	"gossenger/frontend"
	"log"
	"net/http"
)

func main() {
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=postgres dbname=messenger password=password sslmode=disable")
	upgrader := websocket.Upgrader{}
	connectionKeeper := common.UpgradeReminder{}
	defer connectionKeeper.Close()

	if err != nil {
		panic(err)
	}
	defer db.Close()
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	api := dbapi.Api{
		Db:    db,
		Redis: client,
	}

	api.Init()
	var templateManager frontend.TemplateManager
	templateManager.Init("frontend/templates")

	fs := http.FileServer(http.Dir("./frontend/static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/register", backend.RegisterHandler(&api))

	http.HandleFunc("/login", backend.LoginHandler(&api))

	http.HandleFunc("/chats/create", backend.CreateChatHandler(&api))
	http.HandleFunc("/chats/add_user", backend.AddUserToChatHandler(&api))
	http.HandleFunc("/chats/list", backend.ListUserChatsHandler(&api))
	http.HandleFunc("/chats/list_members", backend.ListChatMembersHandler(&api))

	http.HandleFunc("/messages/send", backend.SendMessageHandler(&api, connectionKeeper))
	http.HandleFunc("/messages/list", backend.ListMessagesHandler(&api))

	http.HandleFunc("/login_page", frontend.LoginPageHandler(&api, &templateManager))
	http.HandleFunc("/register_page", frontend.RegisterPageHandler(&api, &templateManager))
	http.HandleFunc("/", frontend.IndexHandler(&api, &templateManager))

	http.HandleFunc("/messages_ws", backend.MessagesHandler(&api, &upgrader, connectionKeeper))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
