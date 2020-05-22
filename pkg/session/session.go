package session

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/interaction"
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
	QuestionsList    []*question.Question
	SelectedQuestion *question.Question
	Participants     map[string]string

	CurrentPhase        Phase
	VotePhaseDeadline   time.Duration
	AnswerPhaseDeadline time.Duration
	ExpiresAt           time.Time

	slack *slack.Client
}

func (m *Manager) NewIcebreaker(ctx context.Context, team *team.Team, questions []*question.Question, channelID string) (*Session, error) {
	duration := time.Duration(team.SessionDurationMins) * time.Minute
	session := &Session{
		ID:                  uuid.New(),
		Team:                team,
		QuestionsList:       questions,
		CurrentPhase:        PhaseNone,
		VotePhaseDeadline:   duration / 2,
		AnswerPhaseDeadline: duration,
		ExpiresAt:           time.Now().Add(duration).Add(5 * time.Second),
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

	err := m.ManageSession(logger, session.Team.TeamID, session.ID.String(), m.startVotePhase())
	if err != nil {
		return err
	}

	time.AfterFunc(session.VotePhaseDeadline, func() {
		err = m.ManageSession(logger, session.Team.TeamID, session.ID.String(), m.startAnswerPhase())
		if err != nil {
			m.logger.Log("event", "session.failed")
			m.TeardownSession(logger, session)
		}
	})

	time.AfterFunc(session.AnswerPhaseDeadline, func() {
		defer m.logger.Log("event", "session.finished")

		err = m.ManageSession(logger, session.Team.TeamID, session.ID.String(), m.startResultsPhase())
		if err != nil {
			m.logger.Log("event", "session.failed")
			m.TeardownSession(logger, session)
		}

		// TODO: save session data
	})

	return nil
}

func (m *Manager) HandleInteractionAction(teamID string, action *interaction.Payload) error {
	logger := log.With(m.logger, "session", action.SessionID, "block", action.BlockID, "value", action.Value)
	switch action.BlockID {
	case interaction.ActionQuestionView:
		err := m.ManageSession(logger, teamID, action.SessionID.String(), m.openQuestionModal(action.TriggerID))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) HandleInteractionResponse(teamID string, response *interaction.Payload) error {
	logger := log.With(m.logger, "session", response.SessionID, "block", response.BlockID, "value", response.Value)
	switch response.BlockID {
	case interaction.ResponseAnswerSubmit:
		err := m.ManageSession(logger, teamID, response.SessionID.String(), m.handleAnswerInput(response))
		if err != nil {
			return err
		}
	}
	return nil
}
