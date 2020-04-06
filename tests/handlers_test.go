package tests

import (
	"bytes"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"io/ioutil"
	"messenger/backend"
	"messenger/dbapi"
	"messenger/frontend"
	"messenger/models"
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
	registerTestHandler := http.HandlerFunc(backend.RegisterHandler(api))

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
	indexTestHandler := http.HandlerFunc(frontend.IndexHandler(api))
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

	handler := http.HandlerFunc(backend.LoginHandler(api))

	loginResponseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(loginResponseRecorder, loginRequest)

	if loginResponseRecorder.Code != http.StatusOK {
		t.Errorf("Login handler returned unexpected status: %v", loginResponseRecorder.Code)
	}
}

func getUserData(name string) *models.RegisterUserRequestSchema {
	return &models.RegisterUserRequestSchema{
		Nickname:  name,
		FirstName: name,
		LastName:  name,
		Password:  name,
	}
}

func createChatInTestEnv(t *testing.T, api *dbapi.Api) []*http.Cookie {
	userA, userB, userC := getUserData("a"), getUserData("b"), getUserData("c")
	usersToRegister := []*models.RegisterUserRequestSchema{userA, userB, userC}

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

	handler := http.HandlerFunc(backend.CreateChatHandler(api))

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

	handler := http.HandlerFunc(backend.AddUserToChatHandler(api))

	addUserToChatResponseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(addUserToChatResponseRecorder, addUserToChatRequest)

	if addUserToChatResponseRecorder.Code != http.StatusOK {
		t.Errorf("Adding user to chat handler returned unexpected status: %v", addUserToChatResponseRecorder.Code)
	}

	chat, err := api.GetChat(1)
	if err != nil {
		t.Fatal(err)
	}
	actualMembers, err := api.ListChatMembers(chat)
	if err != nil {
		t.Fatal(err)
	}

	if len(actualMembers) != 4 {
		t.Errorf("Wrong count of actualMembers: %v", len(actualMembers))
	}

	expectedMemberNames := []string{"a", "b", "c", "newuser"}

	for i := 0; i < len(actualMembers); i++ {
		if actualMembers[i].LastName != actualMembers[i].FirstName ||
			actualMembers[i].FirstName != actualMembers[i].Nickname ||
			actualMembers[i].Nickname != expectedMemberNames[i] {
			t.Fatalf("Wrong member with index %v: actual %v, expected name: %v", i, actualMembers[i], expectedMemberNames[i])
		}
	}
}

func TestSendingMessage(t *testing.T) {
	api := getTestApi()
	defer api.Close()

	adminCookies := createChatInTestEnv(t, api)

	sendMessageRequestContent, err := ioutil.ReadFile("requests/send_message.json")
	if err != nil {
		t.Fatal(err)
	}

	sendMessageRequest, err := http.NewRequest(http.MethodPost, "/messages/send", bytes.NewReader(sendMessageRequestContent))
	if err != nil {
		t.Fatal(err)
	}

	addAllCookies(adminCookies, sendMessageRequest)

	handler := http.HandlerFunc(backend.SendMessageHandler(api))

	sendMessageResponseRecorder := httptest.NewRecorder()
	handler.ServeHTTP(sendMessageResponseRecorder, sendMessageRequest)

	if sendMessageResponseRecorder.Code != http.StatusOK {
		t.Errorf("Sending message handler returned unexpected status: %v", sendMessageResponseRecorder.Code)
	}

	messages, err := api.ListMessages(1)
	if err != nil {
		t.Fatal(err)
	}

	if len(messages) != 1 {
		t.Errorf("Wrong number of messages: %v", len(messages))
	}

	expectedText := "Hello there folks!"

	if messages[0].Text != expectedText {
		t.Errorf("Wrong message text: expected %v, actual message: %v", expectedText, messages[0].Text)
	}
}
