package bingsoo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/postgres"
	"github.com/jace-ys/bingsoo/pkg/question"
	"github.com/jace-ys/bingsoo/pkg/redis"
	"github.com/jace-ys/bingsoo/pkg/session"
	"github.com/jace-ys/bingsoo/pkg/team"
	"github.com/jace-ys/bingsoo/pkg/worker"
)

type BingsooBotConfig struct {
	SigningSecret string
}

type BingsooBot struct {
	logger   log.Logger
	server   *http.Server
	worker   *worker.WorkerPool
	team     *team.Registry
	question *question.Bank
	session  *session.Manager
	secret   string
}

func NewBingsooBot(logger log.Logger, postgres *postgres.Client, redis *redis.Client, secret string) *BingsooBot {
	bot := &BingsooBot{
		logger:   logger,
		server:   &http.Server{},
		worker:   worker.NewWorkerPool(),
		team:     team.NewRegistry(postgres),
		question: question.NewBank(postgres),
		session:  session.NewManager(logger, redis),
		secret:   secret,
	}
	bot.server.Handler = bot.handler()
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

func (bot *BingsooBot) handler() http.Handler {
	router := mux.NewRouter()
	v1 := router.PathPrefix("/api/v1").Subrouter()
	v1.Handle("/health", promhttp.Handler()).Methods(http.MethodGet)

	v1commands := v1.PathPrefix("/commands").Subrouter()
	v1commands.HandleFunc("", bot.commands).Methods(http.MethodPost)
	v1commands.Use(bot.verifySignatureMiddleware)

	return router
}

func (bot *BingsooBot) verifySignatureMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := bot.verifySignature(r)
		if err != nil {
			level.Error(bot.logger).Log("event", "signature.verified", "error", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (bot *BingsooBot) verifySignature(r *http.Request) error {
	verifier, err := slack.NewSecretsVerifier(r.Header, bot.secret)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	_, err = verifier.Write(body)
	if err != nil {
		return err
	}

	return verifier.Ensure()
}

func (bot *BingsooBot) sendJSON(w http.ResponseWriter, code int, res interface{}) {
	response, err := json.Marshal(res)
	if err != nil {
		level.Error(bot.logger).Log("event", "response.encoded", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Bingsoo is currently unavailable. Please try again later."))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(response))
}
