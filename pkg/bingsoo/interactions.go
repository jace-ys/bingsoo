package bingsoo

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/interactions"
	"github.com/jace-ys/bingsoo/pkg/session"
	"github.com/jace-ys/bingsoo/pkg/team"
)

func (bot *BingsooBot) interactions(w http.ResponseWriter, r *http.Request) {
	bot.logger.Log("event", "interaction.started")
	defer bot.logger.Log("event", "interaction.finished")

	ctx := r.Context()

	i, err := bot.parseInteraction(r)
	if err != nil {
		return
	}

	it, err := interactions.ParseType(i)
	if err != nil {
		return
	}

	logger := log.With(bot.logger, "team", i.Team.ID, "user", i.User.ID, "channel", i.Channel.ID, "interaction", it)
	logger.Log("event", "interaction.parsed")

	t, err := bot.team.Get(ctx, i.Team.ID)
	if err != nil {
		logger.Log("event", "team.get", "team", i.Team.ID, "error", err)
		switch {
		case errors.Is(err, team.ErrTeamNotFound):
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			return
		}
	}

	switch it {
	case interactions.Action:
		actions := interactions.GetActions(i)
		for _, action := range actions {
			logger.Log("event", "action.parsed", "session", action.SessionID, "block", action.BlockID, "action", action.ActionID, "value", action.Value)
			bot.handleInteractionAction(logger, t, action)
		}
	case interactions.Response:
		responses := interactions.GetResponses(i)
		for _, response := range responses {
			logger.Log("event", "response.parsed", "session", response.SessionID, "block", response.BlockID, "action", response.ActionID, "value", response.Value)
			bot.handleInteractionResponse(logger, t, response)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (bot *BingsooBot) parseInteraction(r *http.Request) (*slack.InteractionCallback, error) {
	payload := r.FormValue("payload")

	var interaction slack.InteractionCallback
	err := json.Unmarshal([]byte(payload), &interaction)
	if err != nil {
		return nil, err
	}

	return &interaction, nil
}

func (bot *BingsooBot) handleInteractionAction(logger log.Logger, team *team.Team, action *interactions.Payload) {
	switch action.BlockID {
	case interactions.ActionQuestionView:
		session := &session.Session{ID: action.SessionID, Team: team}
		bot.session.ManageSession(logger, session, true, bot.session.OpenAnswerModal(action.TriggerID))
	}
}

func (bot *BingsooBot) handleInteractionResponse(logger log.Logger, team *team.Team, response *interactions.Payload) {
	switch response.BlockID {
	case interactions.ResponseAnswerSubmit:
	}
}
