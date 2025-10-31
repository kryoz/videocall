package ws

import (
	"net/http"
)

type ServiceHandlerInterface interface {
	SignalHandler(w http.ResponseWriter, r *http.Request)
}

type API struct {
	processor ServiceHandlerInterface
}

func NewAPI(handlers ServiceHandlerInterface) *API {
	return &API{processor: handlers}
}

func (api *API) RegisterHandlers() {
	http.HandleFunc("/api/signal", func(w http.ResponseWriter, r *http.Request) {
		api.processor.SignalHandler(w, r)
	})
}
