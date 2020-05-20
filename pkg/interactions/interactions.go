package interactions

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

func ParseType(interaction *slack.InteractionCallback) (InteractionType, error) {
	switch {
	case interaction.ActionCallback.BlockActions != nil:
		return Action, nil
	case interaction.View.State != nil:
		return Response, nil
	default:
		return "", ErrInvalidInteraction
	}
}

func GetActions(interaction *slack.InteractionCallback) []*Payload {
	var actions []*Payload
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

func GetResponses(interaction *slack.InteractionCallback) []*Payload {
	var responses []*Payload
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
