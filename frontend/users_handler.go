package frontend

import (
	"gossenger/backend"
	"gossenger/common"
	"gossenger/dbapi"
	"log"
	"net/http"
	"strconv"
)

type UserPageSchema struct {
	ID        uint
	FirstName string
	LastName  string
	Nickname  string
}

func UsersPageHandler(api *dbapi.Api, templateManager *TemplateManager) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := backend.EnsureLogin(api, request)
		if loggedUser == nil {
			backend.Redirect(writer, "/login_page")
			return
		}

		tpl, err := templateManager.GetTemplate("users")
		if err != nil {
			log.Println("Can not return users page: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		limit, offset, err := common.GetLimitAndOffset(request.URL.Query())
		if err != nil {
			log.Println("Can not return users page: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		users, err := api.ListUsers(limit, offset)
		if err != nil {
			log.Println("Can not return users page: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		var usersResponseData []UserPageSchema
		for _, user := range users {
			usersResponseData = append(usersResponseData, UserPageSchema{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Nickname:  user.Nickname,
			})
		}

		loggedUserData := UserPageSchema{
			ID:        loggedUser.ID,
			FirstName: loggedUser.FirstName,
			LastName:  loggedUser.LastName,
			Nickname:  loggedUser.Nickname,
		}

		pageData := struct {
			Logged UserPageSchema
			Users  []UserPageSchema
		}{
			loggedUserData,
			usersResponseData,
		}

		if err := tpl.Execute(writer, pageData); err != nil {
			log.Println("Can not return users page: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}

func UserPageHandler(api *dbapi.Api, templateManager *TemplateManager) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := backend.EnsureLogin(api, request)
		if loggedUser == nil {
			backend.Redirect(writer, "/login_page")
			return
		}

		rawUserId, ok := request.URL.Query()["user_id"]
		if !ok {
			log.Println("Can not show user page: query parameters must contain user_id")
			writer.WriteHeader(400)
			return
		}

		userId, err := strconv.ParseUint(rawUserId[0], 10, 64)
		if err != nil {
			log.Println("Can not return user page: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		user, err := api.GetUserById(uint(userId))
		if err != nil {
			log.Println("Can not return user page: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		tpl, err := templateManager.GetTemplate("user")
		if err != nil {
			log.Println("Can not return user page: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		requestedUserPageData := UserPageSchema{
			ID:        user.ID,
			Nickname:  user.Nickname,
			FirstName: user.FirstName,
			LastName:  user.LastName,
		}

		loggedUserPageData := UserPageSchema{
			ID:        loggedUser.ID,
			Nickname:  loggedUser.Nickname,
			FirstName: loggedUser.FirstName,
			LastName:  loggedUser.LastName,
		}

		usersPageData := struct {
			Requested UserPageSchema
			Logged    UserPageSchema
		}{
			requestedUserPageData,
			loggedUserPageData,
		}

		if err := tpl.Execute(writer, usersPageData); err != nil {
			log.Println("Can not return user page: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
