package frontend

import (
	"gossenger/backend"
	"gossenger/common"
	"gossenger/dbapi"
	"log"
	"net/http"
)

type userPageSchema struct {
	FirstName string
	LastName  string
	Nickname  string
}

func IndexHandler(api *dbapi.Api, templateManager *TemplateManager) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := backend.EnsureLogin(api, writer, request)
		if user == nil {
			backend.Redirect(writer, "/login_page")
			return
		}

		tpl, err := templateManager.GetTemplate("index")
		if err != nil {
			log.Println("Can not return index page: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		userPageData := userPageSchema{
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Nickname:  user.Nickname,
		}

		if err := tpl.Execute(writer, userPageData); err != nil {
			log.Println("Can not return index page: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
