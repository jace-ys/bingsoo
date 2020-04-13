package bingsoo

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/jace-ys/bingsoo/pkg/slack"
	"github.com/jace-ys/bingsoo/pkg/worker"
)

type BingsooBot struct {
	logger log.Logger
	slack  *slack.Handler
	server *http.Server
	worker *worker.WorkerPool
}

func NewBingsooBot(logger log.Logger, slack *slack.Handler, worker *worker.WorkerPool) *BingsooBot {
	bot := &BingsooBot{
		logger: logger,
		slack:  slack,
		server: &http.Server{},
		worker: worker,
	}
	bot.server.Handler = bot.Handler()
	return bot
}

func (bot *BingsooBot) StartServer(port int) error {
	level.Info(bot.logger).Log("event", "server.started", "port", port)
	defer level.Info(bot.logger).Log("event", "server.stopped")
	bot.server.Addr = fmt.Sprintf(":%d", port)
	if err := bot.server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func (bot *BingsooBot) StartWorkers(ctx context.Context, concurrency int) error {
	level.Info(bot.logger).Log("event", "workers.started", "concurrency", concurrency)
	defer level.Info(bot.logger).Log("event", "workers.stopped")
	if err := bot.worker.Process(ctx, concurrency); err != nil {
		return fmt.Errorf("failed to start workers: %w", err)
	}
	return nil
}

func (bot *BingsooBot) Shutdown(ctx context.Context) error {
	if err := bot.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}
	if err := bot.worker.Close(); err != nil {
		return fmt.Errorf("failed to shutdown workers: %w", err)
	}
	return nil
}

func (bot *BingsooBot) Handler() http.Handler {
	router := mux.NewRouter()
	v1 := router.PathPrefix("/api/v1").Subrouter()
	v1.Handle("/health", promhttp.Handler()).Methods(http.MethodGet)

	v1commands := v1.PathPrefix("/commands").Subrouter()
	v1commands.HandleFunc("", bot.Commands).Methods(http.MethodPost)
	v1commands.Use(bot.VerifySignatureMiddleware)

	return router
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
