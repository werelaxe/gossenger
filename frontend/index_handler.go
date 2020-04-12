package frontend

import (
	"gossenger/backend"
	"gossenger/common"
	"gossenger/dbapi"
	"log"
	"net/http"
)

func IndexHandler(api *dbapi.Api, templateManager *TemplateManager) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := backend.EnsureLogin(api, request)
		if loggedUser == nil {
			backend.Redirect(writer, "/login_page")
			return
		}

		tpl, err := templateManager.GetTemplate("index")
		if err != nil {
			log.Println("Can not return index page: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		userPageData := UserPageSchema{
			ID:        loggedUser.ID,
			FirstName: loggedUser.FirstName,
			LastName:  loggedUser.LastName,
			Nickname:  loggedUser.Nickname,
		}

		if err := tpl.Execute(writer, userPageData); err != nil {
			log.Println("Can not return index page: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
