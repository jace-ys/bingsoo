package bingsoo

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/interaction"
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

	logger := log.With(bot.logger, "team", i.Team.ID, "user", i.User.ID, "channel", i.Channel.ID)
	logger.Log("event", "interaction.parsed", "type", i.Type)

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

	switch i.Type {
	case slack.InteractionTypeBlockActions:
		for _, action := range interaction.ParseBlockActions(i) {
			err := bot.session.HandleInteractionAction(t.TeamID, action)
			if err != nil {
				logger.Log("event", "interaction.failed", "error", err)
				return
			}
		}
	case slack.InteractionTypeViewSubmission:
		for _, response := range interaction.ParseViewSubmission(i) {
			err := bot.session.HandleInteractionResponse(t.TeamID, response)
			if err != nil {
				logger.Log("event", "interaction.failed", "error", err)
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (bot *BingsooBot) parseInteraction(r *http.Request) (*slack.InteractionCallback, error) {
	payload := r.FormValue("payload")

	var i slack.InteractionCallback
	err := json.Unmarshal([]byte(payload), &i)
	if err != nil {
		return nil, err
	}

	return &i, nil
}
