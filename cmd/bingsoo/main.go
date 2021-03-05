package main

import (
	"context"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/jace-ys/go-library/postgres"
	"github.com/jace-ys/go-library/redis"
	"golang.org/x/sync/errgroup"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/jace-ys/bingsoo/pkg/bingsoo"
)

var logger log.Logger

func main() {
	c := parseCommand()

	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	redis := redis.NewClient(c.cache.connectionURL)
	postgres, err := postgres.NewClient(c.database.connectionURL)
	if err != nil {
		exit(err)
	}

	bot := bingsoo.NewBingsooBot(logger, &c.bot, postgres, redis)

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
	database    struct {
		connectionURL string
	}
	cache struct {
		connectionURL string
	}
}

func parseCommand() *config {
	var c config

	kingpin.Flag("port", "Port for the Bingsoo server.").Envar("PORT").Default("8080").IntVar(&c.port)
	kingpin.Flag("concurrency", "Number of concurrent workers to process tasks.").Envar("CONCURRENCY").Default("4").IntVar(&c.concurrency)
	kingpin.Flag("signing-secret", "Signing secret for verifying requests from Slack.").Envar("SIGNING_SECRET").Required().StringVar(&c.bot.SigningSecret)
	kingpin.Flag("database-url", "URL for connecting to Postgres.").Envar("DATABASE_URL").Default("postgres://bingsoo:bingsoo@127.0.0.1:5432/bingsoo").StringVar(&c.database.connectionURL)
	kingpin.Flag("redis-url", "URL for connecting to Redis.").Envar("REDIS_URL").Default("redis://127.0.0.1:6379").StringVar(&c.cache.connectionURL)
	kingpin.Parse()

	return &c
}

func exit(err error) {
	logger.Log("event", "app.fatal", "error", err)
	os.Exit(1)
}
