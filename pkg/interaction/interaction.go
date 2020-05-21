package interaction

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/slack-go/slack"
)

var (
	ErrInvalidInteraction = errors.New("could not handle invalid interaction type")
)

type Payload struct {
	SessionID uuid.UUID
	BlockID   string
	ActionID  string
	TriggerID string
	Value     string
}

func ParseBlockActions(interaction *slack.InteractionCallback) []*Payload {
	var actions []*Payload
	if interaction.ActionCallback.BlockActions == nil {
		return actions
	}

	for _, action := range interaction.ActionCallback.BlockActions {
		sessionID := strings.SplitN(action.ActionID, "/", 2)[0]
		actions = append(actions, &Payload{
			SessionID: uuid.MustParse(sessionID),
			BlockID:   action.BlockID,
			ActionID:  action.ActionID,
			TriggerID: interaction.TriggerID,
			Value:     action.Value,
		})
	}
	return actions
}

func ParseViewSubmission(interaction *slack.InteractionCallback) []*Payload {
	var responses []*Payload
	if interaction.View.State == nil {
		return responses
	}

	for blockID, actions := range interaction.View.State.Values {
		for actionID, action := range actions {
			sessionID := strings.SplitN(actionID, "/", 2)[0]
			responses = append(responses, &Payload{
				SessionID: uuid.MustParse(sessionID),
				BlockID:   blockID,
				ActionID:  actionID,
				TriggerID: interaction.TriggerID,
				Value:     action.Value,
			})
		}
	}
	return responses
}
