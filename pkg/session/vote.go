package session

import (
	"context"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/message"
)

func (m *Manager) startVotePhase() ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) (*Session, error) {
		level.Info(logger).Log("event", "phase.started", "phase", "vote")

		if session.CurrentPhase != PhaseNone {
			return session, ErrUnexpectedPhase
		}
		session.CurrentPhase = PhaseVote

		voteMessage := message.VoteBlock(session.ID.String(), session.Questions)
		_, _, err := session.slack.PostMessageContext(ctx, session.Team.ChannelID, slack.MsgOptionBlocks(voteMessage.BlockSet...))
		if err != nil {
			return session, fmt.Errorf("failed to post start message: %w", err)
		}

		spew.Dump(session)
		return session, nil
	}
}
