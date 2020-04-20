package backend

import (
	"encoding/json"
	"gossenger/common"
	"gossenger/dbapi"
	"gossenger/models"
	"log"
	"net/http"
)

func LoginHandler(api *dbapi.Api) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(400)
			return
		}

		var loginUserData models.LoginUserRequestSchema
		if err := json.NewDecoder(request.Body).Decode(&loginUserData); err != nil {
			if e, ok := err.(*json.SyntaxError); ok {
				log.Printf("syntax error at byte offset %d", e.Offset)
			}
			log.Println("Can not login user: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		ok, err := api.IsValidPair(&loginUserData)
		if err != nil {
			log.Println("Can not check pair: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		if !ok {
			log.Println("Invalid login/password")
			writer.WriteHeader(400)
			return
		}

		salt, _ := api.CreateSession(loginUserData.Nickname)

		if err := Auth(loginUserData.Nickname, writer, salt); err != nil {
			log.Println("Can not login user: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
