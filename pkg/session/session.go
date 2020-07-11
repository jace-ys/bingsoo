package session

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/interaction"
	"github.com/jace-ys/bingsoo/pkg/message"
	"github.com/jace-ys/bingsoo/pkg/question"
	"github.com/jace-ys/bingsoo/pkg/team"
)

type Phase int

const (
	PhaseNone Phase = iota
	PhaseVote
	PhaseAnswer
	PhaseResult
)

type Session struct {
	ID               uuid.UUID
	Team             *team.Team
	Questions        question.QuestionSet
	SelectedQuestion string
	Participants     map[string]string

	CurrentPhase        Phase
	VotePhaseDeadline   time.Duration
	AnswerPhaseDeadline time.Duration
	ExpiresAt           time.Time

	VoteMessage *slack.Msg

	slack *slack.Client
}

func (m *Manager) NewIcebreaker(ctx context.Context, team *team.Team, questions question.QuestionSet, channelID string) (*Session, error) {
	duration := time.Duration(team.SessionDurationMins) * time.Minute
	session := &Session{
		ID:                  uuid.New(),
		Team:                team,
		Questions:           questions,
		CurrentPhase:        PhaseNone,
		VotePhaseDeadline:   duration / 2,
		AnswerPhaseDeadline: duration / 2,
		ExpiresAt:           time.Now().Add(duration).Add(30 * time.Second),
	}

	err := m.validateSession(session, channelID)
	if err != nil {
		return nil, err
	}

	err = m.initSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (m *Manager) StartSession(ctx context.Context, session *Session) error {
	logger := log.With(m.logger, "session", session.ID, "team", session.Team.TeamID, "channel", session.Team.ChannelID)
	logger.Log("event", "session.started")

	err := m.ManageSession(ctx, logger, session.Team.TeamID, session.ID.String(), m.startVotePhase())
	if err != nil {
		return err
	}

	time.AfterFunc(session.VotePhaseDeadline, func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = m.ManageSession(ctx, logger, session.Team.TeamID, session.ID.String(), m.startAnswerPhase())
		if err != nil {
			logger.Log("event", "session.teardown", "error", err)

			err = m.TeardownSession(ctx, session)
			if err != nil {
				logger.Log("event", "session.teardown.failed", "error", err)
			}

			return
		}

		time.AfterFunc(session.AnswerPhaseDeadline, func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err = m.ManageSession(ctx, logger, session.Team.TeamID, session.ID.String(), m.startResultsPhase())
			if err != nil {
				logger.Log("event", "session.teardown", "error", err)

				err = m.TeardownSession(ctx, session)
				if err != nil {
					logger.Log("event", "session.teardown.failed", "error", err)
				}

				return
			}

			logger.Log("event", "session.finished")

			_, err := m.saveSession(ctx, session.Team.TeamID)
			if err != nil {
				logger.Log("event", "session.save.failed", "error", err)
				return
			}

			logger.Log("event", "session.saved")
		})
	})

	return nil
}

func (m *Manager) TeardownSession(ctx context.Context, session *Session) error {
	err := m.deleteSession(ctx, session.Team.TeamID)
	if err != nil {
		return err
	}

	session.slack = slack.New(session.Team.AccessToken)

	errorMessage := message.ErrorMessage()
	_, _, err = session.slack.PostMessageContext(ctx, session.Team.ChannelID, errorMessage)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) HandleInteractionAction(ctx context.Context, teamID string, action *interaction.Payload) error {
	logger := log.With(m.logger, "session", action.SessionID, "block", action.BlockID, "value", action.Value)
	switch action.ActionID {
	case interaction.ActionVoteSubmit:
		err := m.ManageSession(ctx, logger, teamID, action.SessionID.String(), m.handleVoteInput(action))
		if err != nil {
			return err
		}
	case interaction.ActionSuggestionView:
		err := m.ManageSession(ctx, logger, teamID, action.SessionID.String(), m.openSuggestionModal(action))
		if err != nil {
			return err
		}
	case interaction.ActionQuestionView:
		err := m.ManageSession(ctx, logger, teamID, action.SessionID.String(), m.openQuestionModal(action))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) HandleInteractionResponse(ctx context.Context, teamID string, response *interaction.Payload) error {
	logger := log.With(m.logger, "session", response.SessionID, "block", response.BlockID, "value", response.Value)
	switch response.ActionID {
	case interaction.ResponseSuggestionSubmit:
		err := m.ManageSession(ctx, logger, teamID, response.SessionID.String(), m.handleSuggestionInput(response))
		if err != nil {
			return err
		}
	case interaction.ResponseAnswerSubmit:
		err := m.ManageSession(ctx, logger, teamID, response.SessionID.String(), m.handleAnswerInput(response))
		if err != nil {
			return err
		}
	}
	return nil
}
