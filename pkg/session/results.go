package session

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/message"
)

func (m *Manager) startResultsPhase() ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) (*Session, error) {
		logger.Log("event", "phase.started", "phase", "results")

		if session.CurrentPhase != PhaseAnswer {
			return session, fmt.Errorf("%s: %v", ErrUnexpectedPhase, session.CurrentPhase)
		}
		session.CurrentPhase = PhaseResult

		err := m.releaseResults(ctx, session)
		if err != nil {
			return session, err
		}

		// spew.Dump(session)
		return session, nil
	}
}

func (m *Manager) releaseResults(ctx context.Context, session *Session) error {
	resultMessage := message.ResultBlock(session.SelectedQuestion)
	_, _, err := session.slack.PostMessageContext(ctx, session.Team.ChannelID, slack.MsgOptionBlocks(resultMessage.BlockSet...))
	if err != nil {
		return err
	}

	return nil
}
