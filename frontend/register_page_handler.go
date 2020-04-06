package frontend

import (
	"log"
	"messenger/backend"
	"messenger/common"
	"messenger/dbapi"
	"net/http"
)

func RegisterPageHandler(api *dbapi.Api, templateManager *TemplateManager) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := backend.EnsureLogin(api, writer, request)
		if user != nil {
			backend.Redirect(writer, "/")
			return
		}

		tpl, err := templateManager.GetTemplate("register")
		if err != nil {
			log.Println("Can not return register page: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if err := tpl.Execute(writer, struct{}{}); err != nil {
			log.Println("Can not return register page: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
