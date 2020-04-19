package bingsoo

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kit/kit/log/level"
	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/message"
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

	team, err := bot.getTeam(ctx, s.TeamID)
	if err != nil {
		level.Info(bot.logger).Log("event", "team.get", "team", s.TeamID, "error", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch command.action {
	case "help":
		bot.sendJSON(w, http.StatusOK, &slack.Msg{
			ResponseType: slack.ResponseTypeEphemeral,
			Blocks:       message.HelpBlock(team.ChannelID),
		})

	case "start":
		session, err := bot.newIcebreakerSession(team)
		if err != nil {
			level.Info(bot.logger).Log("event", "session.created", "error", err)
			bot.sendJSON(w, http.StatusOK, &slack.Msg{
				ResponseType: slack.ResponseTypeEphemeral,
				Text:         fmt.Sprintf("Hey <@%s>! I'm having some trouble starting the icebreaker session right now. Please try again later.", s.UserID),
			})
			return
		}

		err = session.start(ctx, s.ChannelID)
		if err != nil {
			level.Info(bot.logger).Log("event", "session.failed", "error", err)
			switch {
			case errors.Is(err, ErrInvalidIcebreakersChannel):
				bot.sendJSON(w, http.StatusOK, &slack.Msg{
					ResponseType: slack.ResponseTypeEphemeral,
					Text:         fmt.Sprintf("Hey <@%s>! Icebreaker sessions can only be started in the <#%s> channel.", s.UserID, team.ChannelID),
				})
				return
			default:
				bot.sendJSON(w, http.StatusOK, &slack.Msg{
					ResponseType: slack.ResponseTypeEphemeral,
					Text:         fmt.Sprintf("Hey <@%s>! I'm having some trouble starting the icebreaker session right now. Please try again later.", s.UserID),
				})
				return
			}
		}

		w.WriteHeader(http.StatusOK)

	default:
		bot.sendJSON(w, http.StatusOK, &slack.Msg{
			ResponseType: slack.ResponseTypeEphemeral,
			Blocks:       message.HelpBlock(team.ChannelID),
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
