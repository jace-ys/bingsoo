package session

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/message"
)

func (m *Manager) startResultsPhase() ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) (*Session, error) {
		level.Info(logger).Log("event", "phase.started", "phase", "results")

		if session.CurrentPhase != PhaseAnswer {
			return session, ErrUnexpectedPhase
		}
		session.CurrentPhase = PhaseResult

		err := m.releaseResults(session)
		if err != nil {
			level.Error(logger).Log("event", "session.failed", "error", err)
			return session, err
		}

		spew.Dump(session)
		return session, nil
	}
}

func (m *Manager) releaseResults(session *Session) error {
	resultMessage := message.ResultBlock(session.Questions[0])
	_, _, err := session.slack.PostMessage(session.Team.ChannelID, slack.MsgOptionBlocks(resultMessage.BlockSet...))
	if err != nil {
		return err
	}

	return nil
}
