package session

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/interaction"
	"github.com/jace-ys/bingsoo/pkg/message"
)

func (m *Manager) startVotePhase() ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) error {
		logger.Log("event", "phase.started", "phase", "vote")

		if session.CurrentPhase != PhaseNone {
			return fmt.Errorf("%s: %v", ErrUnexpectedPhase, session.CurrentPhase)
		}
		session.CurrentPhase = PhaseVote

		voteMessage := message.VoteMessage(session.ID.String(), session.Questions)
		channel, timestamp, err := session.slack.PostMessageContext(ctx, session.Team.ChannelID, voteMessage)
		if err != nil {
			return err
		}

		session.VoteMessage = &slack.Msg{Channel: channel, Timestamp: timestamp}

		return nil
	}
}

func (m *Manager) openSuggestionModal(action *interaction.Payload) ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) error {
		logger.Log("event", "modal.opened", "type", "question")

		_, err := session.slack.OpenViewContext(ctx, action.TriggerID, message.SuggestionModal(session.ID.String()))
		if err != nil {
			return err
		}

		return nil
	}
}

func (m *Manager) handleVoteInput(response *interaction.Payload) ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) error {
		logger.Log("event", "input.handled", "type", "vote")

		user, err := session.slack.GetUserInfoContext(ctx, response.User.ID)
		if err != nil {
			return err
		}

		err = session.Questions.AddVote(response.Value, user)
		if err != nil {
			return err
		}

		voteMessage := message.VoteMessage(session.ID.String(), session.Questions)
		channel, timestamp, _, err := session.slack.UpdateMessageContext(ctx, session.VoteMessage.Channel, session.VoteMessage.Timestamp, voteMessage)
		if err != nil {
			return err
		}
		session.VoteMessage = &slack.Msg{Channel: channel, Timestamp: timestamp}

		return nil
	}
}

func (m *Manager) handleSuggestionInput(response *interaction.Payload) ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) error {
		logger.Log("event", "input.handled", "type", "suggestion")

		err := session.Questions.AddQuestion(response.Value)
		if err != nil {
			return err
		}

		voteMessage := message.VoteMessage(session.ID.String(), session.Questions)
		channel, timestamp, _, err := session.slack.UpdateMessageContext(ctx, session.VoteMessage.Channel, session.VoteMessage.Timestamp, voteMessage)
		if err != nil {
			return err
		}

		session.VoteMessage = &slack.Msg{Channel: channel, Timestamp: timestamp}

		return nil
	}
}
