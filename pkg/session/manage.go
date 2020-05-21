package session

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/redis"
)

var (
	ErrUnauthorizedChannel = errors.New("unauthorized icebreakers channel")
	ErrExistingSession     = errors.New("existing session found for team")
	ErrSessionNotFound     = errors.New("session not found")
	ErrUnexpectedPhase     = errors.New("session found in unexpected phase")
)

type Manager struct {
	logger log.Logger
	redis  *redis.Client
}

func NewManager(logger log.Logger, redis *redis.Client) *Manager {
	return &Manager{
		logger: logger,
		redis:  redis,
	}
}

type ManageSessionFunc func(ctx context.Context, logger log.Logger, session *Session) error

func (m *Manager) ManageSession(logger log.Logger, teamID, sessionID string, manage ManageSessionFunc) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session, err := m.retrieveSession(ctx, teamID)
	if err != nil {
		return err
	}

	if session.ID.String() != sessionID {
		return ErrSessionNotFound
	}

	session.slack = slack.New(session.Team.AccessToken)
	err = manage(ctx, logger, session)
	if err != nil {
		return err
	}

	err = m.cacheSession(ctx, session)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) TeardownSession(logger log.Logger, session *Session) {
	logger.Log("event", "session.cleanup")
	// TODO: clean up session in the face of error
}

func (m *Manager) validateSession(session *Session, channelID string) error {
	if channelID != session.Team.ChannelID {
		return ErrUnauthorizedChannel
	}
	return nil
}

func (m *Manager) initSession(ctx context.Context, session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	ttl := session.ExpiresAt.Sub(time.Now()) / time.Second
	err = m.redis.Transact(ctx, func(conn redigo.Conn) error {
		_, err := redigo.String(conn.Do("SET", session.Team.TeamID, string(data), "NX", "EX", strconv.Itoa(int(ttl))))
		return err
	})
	if err != nil {
		switch {
		case errors.Is(err, redigo.ErrNil):
			return ErrExistingSession
		default:
			return err
		}
	}

	return nil
}

func (m *Manager) retrieveSession(ctx context.Context, teamID string) (session *Session, err error) {
	var data []byte
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

	err = json.Unmarshal(data, &session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (m *Manager) cacheSession(ctx context.Context, session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	ttl := session.ExpiresAt.Sub(time.Now()) / time.Second
	err = m.redis.Transact(ctx, func(conn redigo.Conn) error {
		_, err := conn.Do("SET", session.Team.TeamID, string(data), "EX", strconv.Itoa(int(ttl)))
		return err
	})
	if err != nil {
		return err
	}

	return err
}
