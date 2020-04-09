package frontend

import (
	"gossenger/backend"
	"gossenger/common"
	"gossenger/dbapi"
	"log"
	"net/http"
)

func LoginPageHandler(api *dbapi.Api, templateManager *TemplateManager) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := backend.EnsureLogin(api, request)
		if loggedUser != nil {
			backend.Redirect(writer, "/")
			return
		}

		tpl, err := templateManager.GetTemplate("login")
		if err != nil {
			log.Println("Can not return login page: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if err := tpl.Execute(writer, struct{}{}); err != nil {
			log.Println("Can not return login page: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
