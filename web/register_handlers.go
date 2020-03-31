package web

import (
	"encoding/json"
	"log"
	"messenger/dbapi"
	"messenger/models"
	"net/http"
)

func RegisterHandler(api *dbapi.Api) HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			log.Println("Wrong method (should be POST)")
			writer.WriteHeader(400)
			return
		}

		var registerUserData models.RegisterUserSchema
		decoder := json.NewDecoder(request.Body)

		if err := decoder.Decode(&registerUserData); err != nil {
			if e, ok := err.(*json.SyntaxError); ok {
				log.Printf("syntax error at byte offset %d", e.Offset)
			}
			log.Println("Can not register user: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if err := api.RegisterUser(registerUserData.Nickname, registerUserData.FirstName, registerUserData.LastName, registerUserData.Password); err != nil {
			log.Println("Can not register user: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		salt := api.CreateSession(registerUserData.Nickname)
		if err := Auth(registerUserData.Nickname, writer, salt); err != nil {
			log.Println("Can not auth after registration: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
