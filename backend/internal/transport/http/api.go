package http

import (
	"net/http"
	"strings"
)

type ServiceHandlerInterface interface {
	HandleCreateRoom(w http.ResponseWriter, r *http.Request)
	HandleJoinRoom(w http.ResponseWriter, r *http.Request)
	HandleFetchRoom(w http.ResponseWriter, r *http.Request)
	HandleInviteToRoom(w http.ResponseWriter, r *http.Request)
	HandleTurn(w http.ResponseWriter, r *http.Request)
	HandleRegister(w http.ResponseWriter, r *http.Request)
	HandleLogin(w http.ResponseWriter, r *http.Request)
	HandleCreateGuest(w http.ResponseWriter, r *http.Request)
	HandleRefreshToken(w http.ResponseWriter, r *http.Request)
	HandleRevokeToken(w http.ResponseWriter, r *http.Request)
	HandleSubscribePush(w http.ResponseWriter, r *http.Request)
	HandleUnsubscribePush(w http.ResponseWriter, r *http.Request)
	HandleGetVapidPublicKey(w http.ResponseWriter, r *http.Request)
}

type API struct {
	processor ServiceHandlerInterface
}

func NewAPI(handlers ServiceHandlerInterface) *API {
	return &API{processor: handlers}
}

func (api *API) RegisterHandlers() {
	// Auth endpoints
	http.HandleFunc("/api/auth/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		api.processor.HandleRegister(w, r)
	})

	http.HandleFunc("/api/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		api.processor.HandleLogin(w, r)
	})

	http.HandleFunc("/api/auth/guest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		api.processor.HandleCreateGuest(w, r)
	})

	http.HandleFunc("/api/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		api.processor.HandleRefreshToken(w, r)
	})

	http.HandleFunc("/api/auth/revoke", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		api.processor.HandleRevokeToken(w, r)
	})

	// Push notification endpoints
	http.HandleFunc("/api/push/subscribe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		api.processor.HandleSubscribePush(w, r)
	})

	http.HandleFunc("/api/push/unsubscribe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		api.processor.HandleUnsubscribePush(w, r)
	})

	http.HandleFunc("/api/push/vapid-public-key", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		api.processor.HandleGetVapidPublicKey(w, r)
	})

	// Room endpoints
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

		if strings.HasSuffix(r.URL.Path, "/invite") {
			api.processor.HandleInviteToRoom(w, r)
			return
		}

		http.Error(w, "method is not supported yet", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/api/turn", func(w http.ResponseWriter, r *http.Request) {
		api.processor.HandleTurn(w, r)
	})
}
