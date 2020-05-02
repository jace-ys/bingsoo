package session

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
	"github.com/jace-ys/bingsoo/pkg/question"
	"github.com/jace-ys/bingsoo/pkg/redis"
	"github.com/jace-ys/bingsoo/pkg/team"
)

var (
	ErrUnauthorizedChannel = errors.New("unauthorized icebreakers channel")
	ErrExistingSession     = errors.New("existing session found for team")
	ErrSessionNotFound     = errors.New("session not found")
)

type Session struct {
	ID        uuid.UUID
	Team      *team.Team
	Questions []*question.Question
	Duration  time.Duration

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

func (m *Manager) NewIcebreaker(team *team.Team, questions []*question.Question, token string) *Session {
	return &Session{
		ID:        uuid.New(),
		Team:      team,
		Questions: questions,
		Duration:  time.Duration(team.SessionDurationMins) * time.Minute,
		slack:     slack.New(token), // TODO: initialize this on the fly using Team.AccessToken
	}
}

func (m *Manager) StartSession(ctx context.Context, session *Session, channelID string) error {
	level.Info(m.logger).Log("event", "session.started", "session", session.ID, "team", session.Team.TeamID, "domain", session.Team.TeamDomain)

	startMessage := message.StartBlock(session.ID.String(), session.Questions)
	_, _, err := session.slack.PostMessageContext(ctx, session.Team.ChannelID, slack.MsgOptionBlocks(startMessage.BlockSet...))
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	err = m.CacheSession(ctx, session)
	if err != nil {
		// TODO: clean up created session
		return fmt.Errorf("failed to cache session: %w", err)
	}

	time.AfterFunc(session.Duration, func() {
		level.Info(m.logger).Log("event", "session.finished")
	})

	return nil
}

func (m *Manager) ValidateSession(ctx context.Context, session *Session, channelID string) error {
	if channelID != session.Team.ChannelID {
		return ErrUnauthorizedChannel
	}

	// TODO: uncomment to check team has no existing session
	// _, err := m.RetrieveSession(ctx, session.Team.TeamID)
	// if !errors.Is(err, ErrSessionNotFound) {
	// 	return ErrExistingSession
	// }

	return nil
}

func (m *Manager) CacheSession(ctx context.Context, session *Session) error {
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

func (m *Manager) RetrieveSession(ctx context.Context, teamID string) (session *Session, err error) {
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
		return nil, fmt.Errorf("failed to retrieve session: %w", err)
	}

	return session, nil
}
