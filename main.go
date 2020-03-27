package main

import (
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"log"
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

	api := Api{
		db:    db,
		redis: client,
	}

	api.Init()


	http.HandleFunc("/register", registerHandler(&api))
	http.HandleFunc("/login", loginHandler(&api))
	http.HandleFunc("/", indexHandler(&api))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
