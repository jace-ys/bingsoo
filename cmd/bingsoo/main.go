package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/jace-ys/bingsoo/pkg/bingsoo"
	"github.com/jace-ys/bingsoo/pkg/slack"
)

var logger log.Logger

func main() {
	c := parseCommand()

	logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	slack := slack.NewHandler(c.slack.AccessToken, c.slack.SigningSecret)
	bot := bingsoo.NewBingsooBot(logger, slack)

	level.Info(logger).Log("event", "server.started", "port", c.port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", c.port), bot.Handle())
	if err != nil {
		exit(err)
	}
}

type config struct {
	port  int
	slack slack.Config
}

func parseCommand() *config {
	var c config

	kingpin.Flag("port", "Port for the Bingsoo server.").Default("8080").IntVar(&c.port)
	kingpin.Flag("slack-access-token", "Bot user access token for authenticating with the Slack API.").Required().StringVar(&c.slack.AccessToken)
	kingpin.Flag("slack-signing-secret", "Signing secret for verifying requests from Slack.").Required().StringVar(&c.slack.SigningSecret)
	kingpin.Parse()

	return &c
}

func exit(err error) {
	level.Error(logger).Log("event", "server.fatal", "msg", err)
	os.Exit(1)
}
