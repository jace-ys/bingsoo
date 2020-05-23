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
	UserID    string
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
		identifier := strings.SplitN(action.ActionID, "/", 2)

		actions = append(actions, &Payload{
			SessionID: uuid.MustParse(identifier[0]),
			UserID:    interaction.User.ID,
			BlockID:   action.BlockID,
			ActionID:  identifier[1],
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
			identifier := strings.SplitN(actionID, "/", 2)

			responses = append(responses, &Payload{
				SessionID: uuid.MustParse(identifier[0]),
				UserID:    interaction.User.ID,
				BlockID:   blockID,
				ActionID:  identifier[1],
				TriggerID: interaction.TriggerID,
				Value:     action.Value,
			})
		}
	}
	return responses
}
