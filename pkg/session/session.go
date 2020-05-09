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

	Duration     time.Duration
	CurrentPhase Phase

	slack *slack.Client
}
