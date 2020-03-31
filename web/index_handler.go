package web

import (
	"log"
	"messenger/dbapi"
	"net/http"
)

func IndexHandler(api *dbapi.Api) HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		user := EnsureLogin(api, writer, request)
		if user == nil {
			return
		}

		if _, err := writer.Write([]byte("Hello, " + user.FirstName + " " + user.LastName)); err != nil {
			log.Println("Can not write page: " + err.Error())
			writer.WriteHeader(400)
			return
		}

	}
}
