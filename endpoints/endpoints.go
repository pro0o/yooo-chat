package endpoints

import (
	"encoding/json"
	"net/http"

	"log/slog"

	"github.com/gorilla/mux"
	"github.com/pro0o/yoo-chat/auth"
)

type APIServer struct {
	listenAddr string
}

type ApiError struct {
	Error string `json:"error"`
}

func NewAPIServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
	}
}

func WriteJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

func (h *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/yoo-chat", auth.Authenticate(h.handleWebSocket))

	slog.Info("Registered endpoints:")
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			slog.Info("Endpoint", "path", pathTemplate)
		}
		return nil
	})

	err := http.ListenAndServe(h.listenAddr, router)
	if err != nil {
		slog.Error("Failed to start server", "err", err)
	}
}
