package backend

import (
	"encoding/json"
	"gossenger/common"
	"gossenger/dbapi"
	"gossenger/models"
	"log"
	"net/http"
	"strconv"
)

func ListUsersHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := EnsureLogin(api, request)
		if loggedUser == nil {
			writer.WriteHeader(400)
			return
		}

		users, err := api.ListUsers()
		if err != nil {
			log.Println("Can not list users: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		listUsersResponseData := models.ListUsersResponseSchema{}
		for _, user := range users {
			newUserResponse := models.UserResponseSchema{
				ID:        user.ID,
				Nickname:  user.Nickname,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			}
			listUsersResponseData = append(listUsersResponseData, newUserResponse)
		}

		rawListUsersResponseData, err := json.Marshal(listUsersResponseData)
		if err != nil {
			log.Println("Can not list users: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if _, err = writer.Write(rawListUsersResponseData); err != nil {
			log.Println("Can not list users: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}

func ShowUserHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := EnsureLogin(api, request)
		if loggedUser == nil {
			writer.WriteHeader(400)
			return
		}

		rawUserId, ok := request.URL.Query()["user_id"]
		if !ok {
			log.Println("Can not show user: query parameters must contain user_id")
			writer.WriteHeader(400)
			return
		}

		userId, err := strconv.ParseUint(rawUserId[0], 10, 64)
		if err != nil {
			log.Println("Can not show user: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		user, err := api.GetUserById(uint(userId))
		if err != nil {
			log.Println("Can not show user: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		userResponseData := models.UserResponseSchema{
			ID:        user.ID,
			Nickname:  user.Nickname,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		}

		rawUserResponseData, err := json.Marshal(userResponseData)
		if err != nil {
			log.Println("Can not show user: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if _, err = writer.Write(rawUserResponseData); err != nil {
			log.Println("Can not show user: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
