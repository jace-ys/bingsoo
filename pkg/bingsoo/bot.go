package bingsoo

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type BingsooBot struct {
	logger  log.Logger
	handler *mux.Router
}

func NewBingsooBot(logger log.Logger) *BingsooBot {
	return &BingsooBot{
		logger:  logger,
		handler: mux.NewRouter(),
	}
}

func (bot *BingsooBot) Handle() http.Handler {
	bot.handler.HandleFunc("/health", bot.HealthCheck)
	return bot.handler
}

func (bot *BingsooBot) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
