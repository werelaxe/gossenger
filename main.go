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
	connectionKeeper := common.ConnectionKeeper{}
	connectionKeeper.Init()
	defer connectionKeeper.Close()

	if err != nil {
		panic("Database connection error:" + err.Error())
	}
	defer db.Close()
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if status := client.Ping(); status.Err() != nil {
		panic("Redis connection error:" + status.Err().Error())
	}

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

	http.HandleFunc("/chats/create", backend.CreateChatHandler(&api, connectionKeeper))
	http.HandleFunc("/chats/add_user", backend.AddUserToChatHandler(&api))
	http.HandleFunc("/chats/list", backend.ListChatsHandler(&api))
	http.HandleFunc("/chats/list_members", backend.ListChatMembersHandler(&api))

	http.HandleFunc("/messages/send", backend.SendMessageHandler(&api, connectionKeeper))
	http.HandleFunc("/messages/list", backend.ListMessagesHandler(&api))

	http.HandleFunc("/users/list", backend.ListUsersHandler(&api))
	http.HandleFunc("/users/search", backend.SearchUsersHandler(&api))
	http.HandleFunc("/users/show", backend.ShowUserHandler(&api))

	http.HandleFunc("/login_page", frontend.LoginPageHandler(&api, &templateManager))
	http.HandleFunc("/register_page", frontend.RegisterPageHandler(&api, &templateManager))
	http.HandleFunc("/users_page", frontend.UsersPageHandler(&api, &templateManager))
	http.HandleFunc("/user_page", frontend.UserPageHandler(&api, &templateManager))
	http.HandleFunc("/", frontend.IndexHandler(&api, &templateManager))

	http.HandleFunc("/messages_ws", backend.WebSocketHandler(&api, &upgrader, connectionKeeper, common.MessagesConnType))
	http.HandleFunc("/chats_ws", backend.WebSocketHandler(&api, &upgrader, connectionKeeper, common.ChatsConnType))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
