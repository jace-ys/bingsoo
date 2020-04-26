package bingsoo

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kit/kit/log/level"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/icebreaker"
	"github.com/jace-ys/bingsoo/pkg/message"
	"github.com/jace-ys/bingsoo/pkg/team"
)

type command struct {
	action     string
	parameters []string
}

func (bot *BingsooBot) commands(w http.ResponseWriter, r *http.Request) {
	level.Info(bot.logger).Log("event", "command.started")
	defer level.Info(bot.logger).Log("event", "command.finished")

	ctx := r.Context()

	s, err := slack.SlashCommandParse(r)
	if err != nil {
		return
	}

	command := bot.parseCommandText(s.Text)
	level.Info(bot.logger).Log("event", "command.parsed", "action", command.action)

	t, err := bot.team.Get(ctx, s.TeamID)
	if err != nil {
		level.Info(bot.logger).Log("event", "team.get", "team", s.TeamID, "error", err)
		switch {
		case errors.Is(err, team.ErrTeamNotFound):
			w.WriteHeader(http.StatusUnauthorized)
			return
		default:
			bot.defaultError(w, s.UserID)
			return
		}
		return
	}

	switch command.action {
	case "help":
		bot.sendJSON(w, http.StatusOK, &slack.Msg{
			ResponseType: slack.ResponseTypeEphemeral,
			Blocks:       message.HelpBlock(t.ChannelID),
		})

	case "start":
		session, err := bot.session.NewSession(t, bot.token)
		if err != nil {
			level.Info(bot.logger).Log("event", "session.created", "error", err)
			bot.defaultError(w, s.UserID)
			return
		}

		err = bot.session.ValidateSession(ctx, session, s.ChannelID)
		if err != nil {
			level.Info(bot.logger).Log("event", "session.failed", "error", err)
			switch {
			case errors.Is(err, icebreaker.ErrUnauthorizedChannel):
				bot.sendJSON(w, http.StatusOK, &slack.Msg{
					ResponseType: slack.ResponseTypeEphemeral,
					Text:         fmt.Sprintf("Hey <@%s>! Icebreaker sessions can only be started in the <#%s> channel.", s.UserID, t.ChannelID),
				})
				return
			case errors.Is(err, icebreaker.ErrExistingSession):
				bot.sendJSON(w, http.StatusOK, &slack.Msg{
					ResponseType: slack.ResponseTypeEphemeral,
					Text:         fmt.Sprintf("Hey <@%s>! An icebreaker session is already on-going in the <#%s> channel.", s.UserID, t.ChannelID),
				})
				return
			default:
				bot.defaultError(w, s.UserID)
				return
			}
		}

		err = bot.session.StartSession(ctx, session, s.ChannelID)
		if err != nil {
			level.Info(bot.logger).Log("event", "session.failed", "error", err)
			bot.defaultError(w, s.UserID)
			return
		}

		w.WriteHeader(http.StatusOK)

	default:
		bot.sendJSON(w, http.StatusOK, &slack.Msg{
			ResponseType: slack.ResponseTypeEphemeral,
			Blocks:       message.HelpBlock(t.ChannelID),
		})
	}
}

func (bot *BingsooBot) parseCommandText(text string) *command {
	split := strings.Split(text, " ")
	return &command{
		action:     split[0],
		parameters: split[1:],
	}
}

func (bot *BingsooBot) defaultError(w http.ResponseWriter, userID string) {
	bot.sendJSON(w, http.StatusOK, &slack.Msg{
		ResponseType: slack.ResponseTypeEphemeral,
		Text:         fmt.Sprintf("Hey <@%s>! I'm having some trouble starting the icebreaker session right now. Please try again later.", userID),
	})
}
