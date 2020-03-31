package main

import (
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
	"messenger/dbapi"
	"messenger/web"
	"net/http"
)

func main() {
	db, err := gorm.Open("postgres", "host=localhost port=5432 user=postgres dbname=messenger password=password sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	api := dbapi.Api{
		Db:    db,
		Redis: client,
	}

	api.Init()

	http.HandleFunc("/register", web.RegisterHandler(&api))

	http.HandleFunc("/login", web.LoginHandler(&api))

	http.HandleFunc("/chats/create", web.CreateChatHandler(&api))
	http.HandleFunc("/chats/add_user", web.AddUserToChatHandler(&api))

	http.HandleFunc("/messages/send", web.SendMessageHandler(&api))

	http.HandleFunc("/", web.IndexHandler(&api))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
