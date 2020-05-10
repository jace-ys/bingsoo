package session

import (
	"time"

	"github.com/google/uuid"
	"github.com/slack-go/slack"

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

func (m *Manager) NewIcebreaker(team *team.Team, questions []*question.Question) *Session {
	duration := time.Duration(team.SessionDurationMins) * time.Minute
	return &Session{
		ID:                  uuid.New(),
		Team:                team,
		QuestionsList:       questions,
		CurrentPhase:        PhaseNone,
		VotePhaseDeadline:   duration / 2,
		AnswerPhaseDeadline: duration,
		ExpiresAt:           time.Now().Add(duration).Add(5 * time.Second),
	}
}
