package icebreaker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/message"
	"github.com/jace-ys/bingsoo/pkg/redis"
	"github.com/jace-ys/bingsoo/pkg/team"
)

var (
	ErrUnauthorizedChannel = errors.New("unauthorized icebreakers channel")
	ErrExistingSession     = errors.New("existing session found for team")
	ErrSessionNotFound     = errors.New("session not found")
)

type Session struct {
	ID       uuid.UUID
	Team     *team.Team
	Duration time.Duration
	slack    *slack.Client
}

type SessionManager struct {
	logger log.Logger
	redis  *redis.Client
}

func NewSessionManager(logger log.Logger, redis *redis.Client) *SessionManager {
	return &SessionManager{
		logger: logger,
		redis:  redis,
	}
}

func (m *SessionManager) NewSession(team *team.Team, token string) (*Session, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:       id,
		Team:     team,
		Duration: time.Duration(team.SessionDurationMins) * time.Minute,
		slack:    slack.New(token), // TODO: initialize this on the fly using Team.AccessToken
	}, nil
}

func (m *SessionManager) StartSession(ctx context.Context, session *Session, channelID string) error {
	level.Info(m.logger).Log("event", "session.started", "session", session.ID, "team", session.Team.TeamID, "domain", session.Team.TeamDomain)

	startMessage := slack.MsgOptionBlocks(message.StartBlock(session.ID.String()).BlockSet...)
	_, _, err := session.slack.PostMessageContext(ctx, session.Team.ChannelID, startMessage)
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	err = m.Cache(ctx, session)
	if err != nil {
		// TODO: clean up created session
		return fmt.Errorf("failed to cache session: %w", err)
	}

	time.AfterFunc(session.Duration, func() {
		level.Info(m.logger).Log("event", "session.finished")
	})

	return nil
}

func (m *SessionManager) ValidateSession(ctx context.Context, session *Session, channelID string) error {
	if channelID != session.Team.ChannelID {
		return ErrUnauthorizedChannel
	}

	// TODO: uncomment to check team has no existing session
	// _, err := m.Retrieve(ctx, session.Team.TeamID)
	// if err != nil {
	// 	return nil
	// }

	return ErrExistingSession
}

func (m *SessionManager) Cache(ctx context.Context, session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	deadline := time.Duration(1) * time.Minute / time.Second

	err = m.redis.Transact(ctx, func(conn redigo.Conn) error {
		_, err := conn.Do("SET", session.Team.TeamID, string(data), "EX", strconv.Itoa(int(deadline)))
		return err
	})
	if err != nil {
		return err
	}

	return err
}

func (m *SessionManager) Retrieve(ctx context.Context, teamID string) (*Session, error) {
	var data []byte
	var err error
	err = m.redis.Transact(ctx, func(conn redigo.Conn) error {
		data, err = redigo.Bytes(conn.Do("GET", teamID))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, redigo.ErrNil):
			return nil, ErrSessionNotFound
		default:
			return nil, err
		}
	}

	var session Session
	err = json.Unmarshal(data, &session)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session: %w", err)
	}

	return &session, nil
}
