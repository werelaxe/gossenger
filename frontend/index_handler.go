package frontend

import (
	"gossenger/backend"
	"gossenger/common"
	"gossenger/dbapi"
	"log"
	"net/http"
	"strconv"
)

func IndexHandler(api *dbapi.Api, templateManager *TemplateManager) common.HandlerFuncType {
	return func(writer http.ResponseWriter, request *http.Request) {
		loggedUser := backend.EnsureLogin(api, request)
		if loggedUser == nil {
			backend.Redirect(writer, "/login_page")
			return
		}

		var privateUserId uint64
		rawPrivateChatId, ok := request.URL.Query()["ensure_private_chat"]
		if ok {
			var err error
			privateUserId, err = strconv.ParseUint(rawPrivateChatId[0], 10, 64)
			if err != nil {
				log.Printf("Invalid private chat id for ensuring: '%v'\n", rawPrivateChatId)
			} else {

			}
		}

		tpl, err := templateManager.GetTemplate("index")
		if err != nil {
			log.Println("Can not return index page: " + err.Error())
			writer.WriteHeader(400)
			return
		}

		userPageData := UserPageSchema{
			ID:            loggedUser.ID,
			FirstName:     loggedUser.FirstName,
			LastName:      loggedUser.LastName,
			Nickname:      loggedUser.Nickname,
			PrivateUserId: uint(privateUserId),
		}

		if err := tpl.Execute(writer, userPageData); err != nil {
			log.Println("Can not return index page: " + err.Error())
			writer.WriteHeader(400)
			return
		}
	}
}
