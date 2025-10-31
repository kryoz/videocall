package http

import (
	"net/http"
	"strings"
)

type ServiceHandlerInterface interface {
	HandleCreateRoom(w http.ResponseWriter, r *http.Request)
	HandleJoinRoom(w http.ResponseWriter, r *http.Request)
	HandleFetchRoom(w http.ResponseWriter, r *http.Request)
	HandleTurn(w http.ResponseWriter, r *http.Request)
}

type API struct {
	processor ServiceHandlerInterface
}

func NewAPI(handlers ServiceHandlerInterface) *API {
	return &API{processor: handlers}
}

func (api *API) RegisterHandlers() {
	http.HandleFunc("/api/rooms", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		api.processor.HandleCreateRoom(w, r)
	})

	http.HandleFunc("/api/rooms/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if strings.HasSuffix(r.URL.Path, "/fetch") {
			api.processor.HandleFetchRoom(w, r)
			return
		}

		if strings.HasSuffix(r.URL.Path, "/join") {
			api.processor.HandleJoinRoom(w, r)
			return
		}

		http.Error(w, "method is not supported yet", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/turn", func(w http.ResponseWriter, r *http.Request) {
		api.processor.HandleTurn(w, r)
	})
}
