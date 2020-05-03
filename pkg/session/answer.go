package session

import (
	"context"
	"math/rand"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func (m *Manager) startAnswerPhase() ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) (*Session, error) {
		level.Info(logger).Log("event", "phase.started", "phase", "answer")

		if session.CurrentPhase != PhaseVote {
			return session, ErrUnexpectedPhase
		}
		session.CurrentPhase = PhaseAnswer

		participants, err := m.selectParticipants(session)
		if err != nil {
			level.Error(logger).Log("event", "session.failed", "error", err)
			return session, err
		}
		session.Participants = participants

		spew.Dump(session)
		return session, nil
	}
}

func (m *Manager) selectParticipants(session *Session) (map[string]string, error) {
	channel, err := session.slack.GetChannelInfo(session.Team.ChannelID)
	if err != nil {
		return nil, err
	}

	quota := session.Team.ParticipantQuota
	if len(channel.Members) < quota {
		quota = len(channel.Members)
	}

	participants := make(map[string]string, quota)
	for len(participants) < quota {
		userID := channel.Members[rand.Intn(len(channel.Members))]
		participants[userID] = ""
	}

	return participants, nil
}
