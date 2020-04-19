package bingsoo

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
	"github.com/slack-go/slack"
)

type icebreakerSession struct {
	logger   log.Logger
	id       *uuid.UUID
	team     *team
	duration time.Duration
	slack    *slack.Client
}

func (bot *BingsooBot) newIcebreakerSession(team *team) (*icebreakerSession, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &icebreakerSession{
		logger:   bot.logger,
		id:       &id,
		team:     team,
		duration: time.Duration(team.SessionDurationMins) * time.Minute,
		slack:    bot.slack.Client, // to replace with slack.New(team.BotToken)
	}, nil
}

func (s *icebreakerSession) start(ctx context.Context, channelID string) error {
	level.Info(s.logger).Log("event", "session.started", "session", s.id, "team", s.team.TeamID, "domain", s.team.TeamDomain)
	err := s.validate(ctx, channelID)
	if err != nil {
		return err
	}

	message := slack.MsgOptionText(":shaved_ice: *Time for some icebreakers!* :shaved_ice:\n*Here's your question*: What's your name?", true)
	_, _, err = s.slack.PostMessage(s.team.ChannelID, message)
	if err != nil {
		return err
	}

	time.AfterFunc(s.duration, func() {
		level.Info(s.logger).Log("event", "session.finished")
	})

	return nil
}

func (s *icebreakerSession) validate(ctx context.Context, channelID string) error {
	if channelID != s.team.ChannelID {
		return ErrInvalidIcebreakersChannel
	}
	// check that the team has no existing session
	return nil
}
