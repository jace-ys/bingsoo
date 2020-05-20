package session

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/log"
)

func (m *Manager) StartSession(ctx context.Context, session *Session, channelID string) error {
	err := m.validateSession(ctx, session, channelID)
	if err != nil {
		return err
	}

	logger := log.With(m.logger, "session", session.ID, "team", session.Team.TeamID, "channel", session.Team.ChannelID)
	logger.Log("event", "session.started")

	m.ManageSession(logger, session, false, m.startVotePhase())

	time.AfterFunc(session.VotePhaseDeadline, func() {
		m.ManageSession(logger, session, true, m.startAnswerPhase())
	})

	time.AfterFunc(session.AnswerPhaseDeadline, func() {
		defer func() {
			m.logger.Log("event", "session.finished")
			// TODO: save session data
		}()

		m.ManageSession(logger, session, true, m.startResultsPhase())
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
