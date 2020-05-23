package session

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/interaction"
	"github.com/jace-ys/bingsoo/pkg/message"
)

var (
	ErrQuestionNotFound = errors.New("question not found")
	ErrQuestionExists   = errors.New("question already exists in set")
)

func (m *Manager) startVotePhase() ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) error {
		logger.Log("event", "phase.started", "phase", "vote")

		if session.CurrentPhase != PhaseNone {
			return fmt.Errorf("%s: %v", ErrUnexpectedPhase, session.CurrentPhase)
		}
		session.CurrentPhase = PhaseVote

		voteMessage := message.VoteBlock(session.ID.String(), session.QuestionSet)
		channel, timestamp, err := session.slack.PostMessageContext(ctx, session.Team.ChannelID, slack.MsgOptionBlocks(voteMessage.BlockSet...))
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

		_, ok := session.QuestionSet[response.Value]
		if !ok {
			return ErrQuestionNotFound
		}
		session.QuestionSet[response.Value]++

		return nil
	}
}

func (m *Manager) handleSuggestionInput(response *interaction.Payload) ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) error {
		logger.Log("event", "input.handled", "type", "suggestion")

		_, ok := session.QuestionSet[response.Value]
		if ok {
			return ErrQuestionExists
		}
		session.QuestionSet[response.Value] = 0

		voteMessage := message.VoteBlock(session.ID.String(), session.QuestionSet)
		channel, timestamp, _, err := session.slack.UpdateMessageContext(ctx, session.VoteMessage.Channel, session.VoteMessage.Timestamp, slack.MsgOptionBlocks(voteMessage.BlockSet...))
		if err != nil {
			return err
		}

		session.VoteMessage = &slack.Msg{Channel: channel, Timestamp: timestamp}

		return nil
	}
}
