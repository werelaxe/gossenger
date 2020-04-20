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

		limit, offset, err := common.GetLimitAndOffset(request.URL.Query())
		if err != nil {
			log.Println("Can not list users: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		users, err := api.ListUsers(limit, offset)
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

		rawUserId, isUserIdPassed := request.URL.Query()["user_id"]
		nickname, isNicknamePassed := request.URL.Query()["nickname"]

		if !(isUserIdPassed || isNicknamePassed) {
			log.Println("Can not show user: query parameters must contain user_id or nickname")
			writer.WriteHeader(400)
			return
		}

		var user *models.User

		if isNicknamePassed {
			var err error
			user, err = api.GetUserByNickname(nickname[0])
			if err != nil {
				log.Println("Can not show user: " + err.Error())
				writer.WriteHeader(400)
				return
			}
		} else {
			userId, err := strconv.ParseUint(rawUserId[0], 10, 64)
			if err != nil {
				log.Println("Can not show user: " + err.Error())
				writer.WriteHeader(400)
				return
			}

			user, err = api.GetUserById(uint(userId))
			if err != nil {
				log.Println("Can not show user: " + err.Error())
				writer.WriteHeader(400)
				return
			}
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

func SearchUsersHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := EnsureLogin(api, request)
		if loggedUser == nil {
			writer.WriteHeader(400)
			return
		}

		filter, ok := request.URL.Query()["filter"]
		if !ok {
			log.Println("Can not search users: query parameters must contain filter")
			writer.WriteHeader(400)
			return
		}

		limit, offset, err := common.GetLimitAndOffset(request.URL.Query())
		if err != nil {
			log.Println("Can not search users: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		users, err := api.SearchUsers(filter[0], limit, offset)
		if err != nil {
			log.Println("Can not search users: " + err.Error())
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
			log.Println("Can not search users: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if _, err = writer.Write(rawListUsersResponseData); err != nil {
			log.Println("Can not search users: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
