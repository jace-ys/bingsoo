package session

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/go-kit/kit/log"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/message"
)

func (m *Manager) startAnswerPhase() ManageSessionFunc {
	return func(ctx context.Context, logger log.Logger, session *Session) (*Session, error) {
		logger.Log("event", "phase.started", "phase", "answer")

		if session.CurrentPhase != PhaseVote {
			return session, fmt.Errorf("%s: %v", ErrUnexpectedPhase, session.CurrentPhase)
		}
		session.CurrentPhase = PhaseAnswer

		participants, err := m.selectParticipants(ctx, session)
		if err != nil {
			return session, err
		}
		session.Participants = participants

		// session.ChosenQuestion = m.selectQuestion(session)

		err = m.deliverQuestion(ctx, session)
		if err != nil {
			return session, err
		}

		// spew.Dump(session)
		return session, nil
	}
}

func (m *Manager) selectParticipants(ctx context.Context, session *Session) (map[string]string, error) {
	params := &slack.GetUsersInConversationParameters{ChannelID: session.Team.ChannelID}
	users, _, err := session.slack.GetUsersInConversationContext(ctx, params)
	if err != nil {
		return nil, err
	}

	// TODO: filter out bot users
	quota := session.Team.ParticipantQuota
	if len(users) < quota {
		quota = len(users)
	}

	participants := make(map[string]string, quota)
	for len(participants) < quota {
		userID := users[rand.Intn(len(users))]
		participants[userID] = ""
	}

	return participants, nil
}

func (m *Manager) deliverQuestion(ctx context.Context, session *Session) error {
	for user := range session.Participants {
		params := &slack.OpenConversationParameters{Users: []string{user}}
		channel, _, _, err := session.slack.OpenConversationContext(ctx, params)
		if err != nil {
			return err
		}

		questionMessage := message.QuestionBlock(session.Team.ChannelID, session.Questions[0])
		_, _, err = session.slack.PostMessageContext(ctx, channel.ID, slack.MsgOptionBlocks(questionMessage.BlockSet...))
		if err != nil {
			return err
		}
	}

	return nil
}
