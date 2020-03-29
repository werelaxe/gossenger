package tests

import (
	"bytes"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io/ioutil"
	"messenger/dbapi"
	"messenger/web"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func getTestApi() *dbapi.Api {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
	})

	client.FlushAll()

	api := dbapi.Api{
		Db:    db,
		Redis: client,
	}

	api.Init()
	return &api
}

func TestRegistration(t *testing.T) {
	fmt.Println(os.Getwd())
	api := getTestApi()
	defer api.Close()

	registerRequestContent, err := ioutil.ReadFile("requestsJson/register.json")
	if err != nil {
		t.Fatal(err)
	}

	registrationRequest, err := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(registerRequestContent))
	if err != nil {
		t.Fatal(err)
	}

	registrationResponseRecorder := httptest.NewRecorder()
	registerTestHandler := http.HandlerFunc(web.RegisterHandler(api))

	registerTestHandler.ServeHTTP(registrationResponseRecorder, registrationRequest)

	if registrationResponseRecorder.Code != http.StatusOK {
		t.Errorf("Register handler returned unexpected status: %v", registrationResponseRecorder.Code)
	}

	cookies := registrationResponseRecorder.Result().Cookies()

	indexRequest, err := http.NewRequest(http.MethodPost, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, cookie := range cookies {
		indexRequest.AddCookie(cookie)
	}

	indexResponseRecorder := httptest.NewRecorder()
	indexTestHandler := http.HandlerFunc(web.IndexHandler(api))
	indexTestHandler.ServeHTTP(indexResponseRecorder, indexRequest)

	if indexResponseRecorder.Code != http.StatusOK {
		t.Errorf("Index handler returned unexpected status: %v", indexResponseRecorder.Code)
	}
}

func TestRegistrationAndLogin(t *testing.T) {
	api := getTestApi()
	defer api.Close()

	registerRequestContent, err := ioutil.ReadFile("requestsJson/register.json")
	if err != nil {
		t.Fatal(err)
	}

	registrationRequest, err := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(registerRequestContent))
	if err != nil {
		t.Fatal(err)
	}

	registrationResponseRecorder := httptest.NewRecorder()
	handler := http.HandlerFunc(web.RegisterHandler(api))

	handler.ServeHTTP(registrationResponseRecorder, registrationRequest)

	if registrationResponseRecorder.Code != http.StatusOK {
		t.Errorf("Register handler returned unexpected status: %v", registrationResponseRecorder.Code)
	}

	loginRequestContent, err := ioutil.ReadFile("requestsJson/login.json")
	if err != nil {
		t.Fatal(err)
	}

	loginRequest, err := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginRequestContent))
	if err != nil {
		t.Fatal(err)
	}

	handler = http.HandlerFunc(web.LoginHandler(api))

	loginResponseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginResponseRecorder, loginRequest)

	if loginResponseRecorder.Code != http.StatusOK {
		t.Errorf("Login handler returned unexpected status: %v", loginResponseRecorder.Code)
	}
}
