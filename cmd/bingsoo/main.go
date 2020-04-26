package main

import (
	"context"
	"os"

	"github.com/jace-ys/bingsoo/pkg/team"

	"github.com/alecthomas/kingpin"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"golang.org/x/sync/errgroup"

	"github.com/jace-ys/bingsoo/pkg/bingsoo"
	"github.com/jace-ys/bingsoo/pkg/postgres"
	"github.com/jace-ys/bingsoo/pkg/worker"
)

var logger log.Logger

func main() {
	c := parseCommand()

	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	postgres, err := postgres.NewClient(c.database.Host, c.database.User, c.database.Password, c.database.Database)
	if err != nil {
		exit(err)
	}

	teams := team.NewRegistry(postgres)
	worker := worker.NewWorkerPool()
	bot := bingsoo.NewBingsooBot(logger, worker, teams, c.bot.SigningSecret, c.bot.AccessToken)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return bot.StartServer(c.port)
	})
	g.Go(func() error {
		return bot.StartWorkers(ctx, c.concurrency)
	})
	g.Go(func() error {
		select {
		case <-ctx.Done():
			if err := bot.Shutdown(ctx); err != nil {
				return err
			}
			return ctx.Err()
		}
	})

	if err := g.Wait(); err != nil {
		exit(err)
	}
}

type config struct {
	port        int
	concurrency int
	bot         bingsoo.BingsooBotConfig
	database    postgres.ClientConfig
}

func parseCommand() *config {
	var c config

	kingpin.Flag("port", "Port for the Bingsoo server.").Default("8080").IntVar(&c.port)
	kingpin.Flag("concurrency", "Number of concurrent workers to process tasks.").Default("4").IntVar(&c.concurrency)
	kingpin.Flag("signing-secret", "Signing secret for verifying requests from Slack.").Required().StringVar(&c.bot.SigningSecret)
	kingpin.Flag("access-token", "Bot user access token for authenticating with the Slack API.").Required().StringVar(&c.bot.AccessToken)
	kingpin.Flag("postgres-host", "Host for connecting to Postgres").Default("127.0.0.1:5432").StringVar(&c.database.Host)
	kingpin.Flag("postgres-user", "User for connecting to Postgres").Default("postgres").StringVar(&c.database.User)
	kingpin.Flag("postgres-password", "Password for connecting to Postgres").Required().StringVar(&c.database.Password)
	kingpin.Flag("postgres-db", "Database for connecting to Postgres").Default("postgres").StringVar(&c.database.Database)
	kingpin.Parse()

	return &c
}

func exit(err error) {
	level.Error(logger).Log("event", "app.fatal", "error", err)
	os.Exit(1)
}
