package session

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	redigo "github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/jace-ys/go-library/postgres"
	"github.com/jace-ys/go-library/redis"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/slack-go/slack"
)

var (
	ErrUnauthorizedChannel = errors.New("unauthorized icebreakers channel")
	ErrSessionExists       = errors.New("existing session found for team")
	ErrSessionNotFound     = errors.New("session not found")
	ErrUnexpectedPhase     = errors.New("session found in unexpected phase")
)

type Manager struct {
	logger   log.Logger
	cache    *redis.Client
	database *postgres.Client
}

func NewManager(logger log.Logger, cache *redis.Client, database *postgres.Client) *Manager {
	return &Manager{
		logger:   logger,
		cache:    cache,
		database: database,
	}
}

type ManageSessionFunc func(ctx context.Context, logger log.Logger, session *Session) error

func (m *Manager) ManageSession(ctx context.Context, logger log.Logger, teamID, sessionID string, manage ManageSessionFunc) error {
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
	err = m.cache.Call(ctx, func(conn redigo.Conn) error {
		_, err := redigo.String(conn.Do("SET", session.Team.TeamID, string(data), "NX", "EX", strconv.Itoa(int(ttl))))
		return err
	})
	if err != nil {
		switch {
		case errors.Is(err, redigo.ErrNil):
			return ErrSessionExists
		default:
			return err
		}
	}

	return nil
}

func (m *Manager) retrieveSession(ctx context.Context, teamID string) (session *Session, err error) {
	var data []byte
	err = m.cache.Call(ctx, func(conn redigo.Conn) error {
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
	err = m.cache.Call(ctx, func(conn redigo.Conn) error {
		_, err := conn.Do("SET", session.Team.TeamID, string(data), "EX", strconv.Itoa(int(ttl)))
		return err
	})
	if err != nil {
		return err
	}

	return err
}

func (m *Manager) deleteSession(ctx context.Context, teamID string) error {
	return m.cache.Call(ctx, func(conn redigo.Conn) error {
		_, err := conn.Do("DEL", teamID)
		return err
	})
}

type SessionEnvelope struct {
	ID               uuid.UUID
	TeamID           string
	SelectedQuestion string
	QuestionVotes    jsonMap
	Responses        jsonMap
}

type jsonMap map[string]interface{}

func (m jsonMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Manager) saveSession(ctx context.Context, teamID string) (string, error) {
	session, err := m.retrieveSession(ctx, teamID)
	if err != nil {
		return "", nil
	}

	envelope := &SessionEnvelope{
		ID:               session.ID,
		TeamID:           session.Team.TeamID,
		SelectedQuestion: session.SelectedQuestion,
		QuestionVotes:    make(jsonMap),
		Responses:        make(jsonMap),
	}

	for question, votes := range session.Questions {
		envelope.QuestionVotes[question] = len(votes)
	}

	for user, response := range session.Participants {
		envelope.Responses[user] = response
	}

	err = m.database.Transact(ctx, func(tx *sqlx.Tx) error {
		query := `
		INSERT INTO sessions (id, team_id, question_votes, selected_question, responses)
		VALUES (:id, :team_id, :question_votes, :selected_question, :responses)
		RETURNING id
		`
		stmt, err := tx.PrepareNamedContext(ctx, query)
		if err != nil {
			return err
		}
		row := stmt.QueryRowxContext(ctx, envelope)
		return row.Scan(&session.ID)
	})
	if err != nil {
		var pqErr *pq.Error
		switch {
		case errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation":
			return "", ErrSessionExists
		default:
			return "", err
		}
	}

	return session.ID.String(), nil
}
