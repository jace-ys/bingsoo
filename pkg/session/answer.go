package session

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/go-kit/kit/log"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/interaction"
	"github.com/jace-ys/bingsoo/pkg/message"
)

var (
	ErrParticipantNotFound = errors.New("participant not found")
)

func (m *Manager) startAnswerPhase() ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) error {
		logger.Log("event", "phase.started", "phase", "answer")

		if session.CurrentPhase != PhaseVote {
			return fmt.Errorf("%s: %v", ErrUnexpectedPhase, session.CurrentPhase)
		}
		session.CurrentPhase = PhaseAnswer

		participants, err := m.selectParticipants(ctx, session)
		if err != nil {
			return err
		}

		session.Participants = participants
		session.SelectedQuestion = m.selectQuestion(session)

		err = m.deliverQuestion(ctx, session)
		if err != nil {
			return err
		}

		return nil
	}
}

func (m *Manager) selectParticipants(ctx context.Context, session *Session) (map[string]string, error) {
	params := &slack.GetUsersInConversationParameters{ChannelID: session.Team.ChannelID}
	userIDs, _, err := session.slack.GetUsersInConversationContext(ctx, params)
	if err != nil {
		return nil, err
	}

	users, err := session.slack.GetUsersInfoContext(ctx, userIDs...)
	if err != nil {
		return nil, err
	}

	var humans []string
	for _, user := range *users {
		if !user.IsBot {
			humans = append(humans, user.ID)
		}
	}

	quota := session.Team.ParticipantQuota
	if len(humans) < quota {
		quota = len(humans)
	}

	participants := make(map[string]string, quota)
	for len(participants) < quota {
		userID := humans[rand.Intn(len(humans))]
		participants[userID] = ""
	}

	return participants, nil
}

func (m *Manager) selectQuestion(session *Session) string {
	var selected string
	highest := -1
	for question, votes := range session.Questions {
		if len(votes) > highest {
			selected = question
			highest = len(votes)
		}
	}
	return selected
}

func (m *Manager) deliverQuestion(ctx context.Context, session *Session) error {
	for user := range session.Participants {
		params := &slack.OpenConversationParameters{Users: []string{user}}
		channel, _, _, err := session.slack.OpenConversationContext(ctx, params)
		if err != nil {
			return err
		}

		questionMessage := message.QuestionMessage(session.ID.String(), session.Team.ChannelID)
		_, _, err = session.slack.PostMessageContext(ctx, channel.ID, questionMessage)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) openQuestionModal(action *interaction.Payload) ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) error {
		logger.Log("event", "modal.opened", "type", "question")

		_, err := session.slack.OpenViewContext(ctx, action.TriggerID, message.AnswerModal(session.ID.String(), session.SelectedQuestion))
		if err != nil {
			return err
		}

		return nil
	}
}

func (m *Manager) handleAnswerInput(response *interaction.Payload) ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) error {
		logger.Log("event", "input.handled", "type", "answer")

		_, ok := session.Participants[response.User.ID]
		if !ok {
			return ErrParticipantNotFound
		}
		session.Participants[response.User.ID] = response.Value

		return nil
	}
}
