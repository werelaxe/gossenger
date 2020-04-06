package common

import "net/http"

type HandlerFuncType func(writer http.ResponseWriter, request *http.Request)
