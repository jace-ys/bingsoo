package icebreaker

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/message"
	"github.com/jace-ys/bingsoo/pkg/team"
)

var (
	ErrUnauthorizedChannel = errors.New("unauthorized icebreakers channel")
)

type Session struct {
	logger   log.Logger
	id       *uuid.UUID
	team     *team.Team
	duration time.Duration
	slack    *slack.Client
}

func NewSession(logger log.Logger, team *team.Team, token string) (*Session, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &Session{
		logger:   logger,
		id:       &id,
		team:     team,
		duration: time.Duration(team.SessionDurationMins) * time.Minute,
		slack:    slack.New(token),
	}, nil
}

func (s *Session) Start(ctx context.Context, channelID string) error {
	level.Info(s.logger).Log("event", "session.started", "session", s.id, "team", s.team.TeamID, "domain", s.team.TeamDomain)
	err := s.validate(ctx, channelID)
	if err != nil {
		return err
	}

	startMessage := slack.MsgOptionBlocks(message.StartBlock(s.id.String()).BlockSet...)
	_, _, err = s.slack.PostMessageContext(ctx, s.team.ChannelID, startMessage)
	if err != nil {
		return err
	}

	time.AfterFunc(s.duration, func() {
		level.Info(s.logger).Log("event", "session.finished")
	})

	return nil
}

func (s *Session) validate(ctx context.Context, channelID string) error {
	if channelID != s.team.ChannelID {
		return ErrUnauthorizedChannel
	}
	// check that the team has no existing session
	return nil
}
