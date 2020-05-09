package session

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/message"
)

func (m *Manager) startVotePhase() ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) (*Session, error) {
		logger.Log("event", "phase.started", "phase", "vote")

		if session.CurrentPhase != PhaseNone {
			return session, fmt.Errorf("%s: %v", ErrUnexpectedPhase, session.CurrentPhase)
		}
		session.CurrentPhase = PhaseVote

		voteMessage := message.VoteBlock(session.ID.String(), session.Questions)
		_, _, err := session.slack.PostMessageContext(ctx, session.Team.ChannelID, slack.MsgOptionBlocks(voteMessage.BlockSet...))
		if err != nil {
			return session, fmt.Errorf("failed to post start message: %w", err)
		}

		// spew.Dump(session)
		return session, nil
	}
}
