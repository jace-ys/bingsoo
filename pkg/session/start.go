package session

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func (m *Manager) StartSession(ctx context.Context, session *Session, channelID string) error {
	logger := log.With(m.logger, "session", session.ID, "team", session.Team.TeamID, "domain", session.Team.TeamDomain)
	level.Info(logger).Log("event", "session.started")

	err := m.validateSession(ctx, session, channelID)
	if err != nil {
		return err
	}

	votePhase := session.Duration / 2
	answerPhase := session.Duration

	m.ManageSession(ctx, logger, session, false, m.startVotePhase())

	time.AfterFunc(votePhase, func() {
		m.ManageSession(ctx, logger, session, true, m.startAnswerPhase())
	})

	time.AfterFunc(answerPhase, func() {
		m.ManageSession(ctx, logger, session, true, m.startResultsPhase())
		level.Info(m.logger).Log("event", "session.finished")
	})

	return nil
}

func (m *Manager) validateSession(ctx context.Context, session *Session, channelID string) error {
	if channelID != session.Team.ChannelID {
		return ErrUnauthorizedChannel
	}

	// TODO: uncomment to check team has no existing session
	session, err := m.retrieveSession(ctx, session.Team.TeamID)
	if err == nil {
		return ErrExistingSession
	}

	if errors.Is(err, ErrSessionNotFound) {
		return nil
	}

	return err
}
