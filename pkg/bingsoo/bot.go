package bingsoo

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/jace-ys/bingsoo/pkg/slack"
)

type BingsooBot struct {
	logger  log.Logger
	slack   *slack.Handler
	handler *mux.Router
}

func NewBingsooBot(logger log.Logger, slack *slack.Handler) *BingsooBot {
	return &BingsooBot{
		logger:  logger,
		slack:   slack,
		handler: mux.NewRouter(),
	}
}

func (bot *BingsooBot) Handle() http.Handler {
	v1 := bot.handler.PathPrefix("/api/v1").Subrouter()
	v1.Handle("/health", promhttp.Handler()).Methods(http.MethodGet)

	v1commands := v1.PathPrefix("/commands").Subrouter()
	v1commands.HandleFunc("", bot.Commands).Methods(http.MethodPost)
	v1commands.Use(bot.VerifySignatureMiddleware)

	return bot.handler
}

func (bot *BingsooBot) VerifySignatureMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := bot.slack.VerifySignature(r)
		if err != nil {
			level.Error(bot.logger).Log("event", "verify_signature.failure", "msg", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
