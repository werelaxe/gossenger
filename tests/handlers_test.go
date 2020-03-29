package tests

import (
	"bytes"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io/ioutil"
	"messenger/dbapi"
	"messenger/models"
	"messenger/web"
	"net/http"
	"net/http/httptest"
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

func registerInTestEnv(t *testing.T, api *dbapi.Api, content []byte) []*http.Cookie {
	var registerRequestContent []byte
	if content != nil {
		registerRequestContent = content
	} else {
		registerRequestContentFromFile, err := ioutil.ReadFile("requests/register.json")
		if err != nil {
			t.Fatal(err)
		}
		registerRequestContent = registerRequestContentFromFile
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

	return registrationResponseRecorder.Result().Cookies()
}

func addAllCookies(cookies []*http.Cookie, r *http.Request) {
	for _, cookie := range cookies {
		r.AddCookie(cookie)
	}
}

func TestRegistration(t *testing.T) {
	api := getTestApi()
	defer api.Close()

	cookies := registerInTestEnv(t, api, nil)

	indexRequest, err := http.NewRequest(http.MethodPost, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	addAllCookies(cookies, indexRequest)

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

	registerInTestEnv(t, api, nil)

	loginRequestContent, err := ioutil.ReadFile("requests/login.json")
	if err != nil {
		t.Fatal(err)
	}

	loginRequest, err := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginRequestContent))
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(web.LoginHandler(api))

	loginResponseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginResponseRecorder, loginRequest)

	if loginResponseRecorder.Code != http.StatusOK {
		t.Errorf("Login handler returned unexpected status: %v", loginResponseRecorder.Code)
	}
}

func getUserData(name string) *models.RegisterUserSchema {
	return &models.RegisterUserSchema{
		Nickname:  name,
		FirstName: name,
		LastName:  name,
		Password:  name,
	}
}

func createChatInTestEnv(t *testing.T, api *dbapi.Api) []*http.Cookie {
	userA, userB, userC := getUserData("a"), getUserData("b"), getUserData("c")
	usersToRegister := []*models.RegisterUserSchema{userA, userB, userC}

	var adminCookies []*http.Cookie
	for _, user := range usersToRegister {
		rawUser, err := json.Marshal(user)
		if err != nil {
			t.Fatal(err)
		}
		adminCookies = registerInTestEnv(t, api, rawUser)
	}

	createChatRequestContent, err := ioutil.ReadFile("requests/create_chat.json")
	if err != nil {
		t.Fatal(err)
	}

	createChatRequest, err := http.NewRequest(http.MethodPost, "/chats/create", bytes.NewReader(createChatRequestContent))
	if err != nil {
		t.Fatal(err)
	}

	addAllCookies(adminCookies, createChatRequest)

	handler := http.HandlerFunc(web.CreateChatHandler(api))

	createChatResponseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(createChatResponseRecorder, createChatRequest)

	if createChatResponseRecorder.Code != http.StatusOK {
		t.Errorf("Create chat handler returned unexpected status: %v", createChatResponseRecorder.Code)
	}

	return adminCookies
}

func TestCreatingChat(t *testing.T) {
	api := getTestApi()
	defer api.Close()

	createChatInTestEnv(t, api)
}

func TestAddingUserToChat(t *testing.T) {
	api := getTestApi()
	defer api.Close()

	adminCookies := createChatInTestEnv(t, api)

	newUser := getUserData("newuser")
	rawNewUser, err := json.Marshal(newUser)
	if err != nil {
		t.Fatal(err)
	}

	registerInTestEnv(t, api, rawNewUser)

	addUserToChatRequestContent, err := ioutil.ReadFile("requests/add_user_to_chat.json")
	if err != nil {
		t.Fatal(err)
	}

	addUserToChatRequest, err := http.NewRequest(http.MethodPost, "/chats/add_user", bytes.NewReader(addUserToChatRequestContent))
	if err != nil {
		t.Fatal(err)
	}

	addAllCookies(adminCookies, addUserToChatRequest)

	handler := http.HandlerFunc(web.AddUserToChatHandler(api))

	addUserToChatResponseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(addUserToChatResponseRecorder, addUserToChatRequest)

	if addUserToChatResponseRecorder.Code != http.StatusOK {
		t.Errorf("Create chat handler returned unexpected status: %v", addUserToChatResponseRecorder.Code)
	}

	chat, err := api.GetChat(1)
	if err != nil {
		t.Fatal(err)
	}
	members, err := api.ListChatMembers(chat)
	if err != nil {
		t.Fatal(err)
	}

	if len(members) != 4 {
		t.Errorf("Wrong count of members: %v", len(members))
	}
}
