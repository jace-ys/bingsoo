package session

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/question"
	"github.com/jace-ys/bingsoo/pkg/redis"
	"github.com/jace-ys/bingsoo/pkg/team"
)

var (
	ErrUnauthorizedChannel = errors.New("unauthorized icebreakers channel")
	ErrExistingSession     = errors.New("existing session found for team")
	ErrSessionNotFound     = errors.New("session not found")
	ErrUnexpectedPhase     = errors.New("session found in unexpected phase")
)

type Phase int

const (
	PhaseNone Phase = iota
	PhaseVote
	PhaseAnswer
	PhaseResult
)

type Session struct {
	ID           uuid.UUID
	Team         *team.Team
	Questions    []*question.Question
	Participants map[string]string

	Duration     time.Duration
	CurrentPhase Phase

	slack *slack.Client
}

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

func (m *Manager) NewIcebreaker(team *team.Team, questions []*question.Question) *Session {
	return &Session{
		ID:           uuid.New(),
		Team:         team,
		Questions:    questions,
		CurrentPhase: PhaseNone,
		Duration:     time.Duration(team.SessionDurationMins) * time.Minute,
	}
}

type ManageSessionFunc func(ctx context.Context, logger log.Logger, session *Session) (*Session, error)

func (m *Manager) ManageSession(ctx context.Context, logger log.Logger, session *Session, existing bool, manage ManageSessionFunc) {
	var err error
	defer func() {
		if err != nil {
			// TODO: clean up session in the face of error
			level.Info(logger).Log("event", "session.cleanup")
		}
	}()

	if existing {
		session, err = m.retrieveSession(ctx, session.Team.TeamID)
		if err != nil {
			level.Error(logger).Log("event", "session.failed", "error", err)
			return
		}
	}

	session.slack = slack.New(session.Team.AccessToken)
	session, err = manage(ctx, logger, session)
	if err != nil {
		level.Error(logger).Log("event", "session.failed", "error", err)
		return
	}

	err = m.cacheSession(ctx, session)
	if err != nil {
		level.Error(logger).Log("event", "session.failed", "error", err)
		return
	}
}

func (m *Manager) cacheSession(ctx context.Context, session *Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	ttl := int(session.Duration / time.Second)
	err = m.redis.Transact(ctx, func(conn redigo.Conn) error {
		_, err := conn.Do("SET", session.Team.TeamID, string(data), "EX", strconv.Itoa(ttl))
		return err
	})
	if err != nil {
		return err
	}

	return err
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
