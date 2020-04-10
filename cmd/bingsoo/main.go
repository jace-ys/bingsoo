package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
)

var logger log.Logger

func main() {
	c := parseCommand()

	logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	router := mux.NewRouter()
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	level.Info(logger).Log("event", "server.started", "port", c.serverPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", c.serverPort), router)
	if err != nil {
		exit(err)
	}
}

type config struct {
	serverPort int
}

func parseCommand() *config {
	var c config

	kingpin.Flag("port", "Port for the Bingsoo server.").Default("8080").IntVar(&c.serverPort)
	kingpin.Parse()

	return &c
}

func exit(err error) {
	level.Error(logger).Log("event", "service.fatal", "msg", err)
	os.Exit(1)
}
