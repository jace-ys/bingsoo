package bingsoo

import (
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

	s, err := slack.SlashCommandParse(r)
	if err != nil {
		return
	}

	command := bot.parseCommandText(s.Text)
	level.Info(bot.logger).Log("event", "command.parsed", "action", command.action)

	switch command.action {
	case "help":
		bot.sendJSON(w, http.StatusOK, &slack.Msg{
			ResponseType: slack.ResponseTypeEphemeral,
			Blocks:       message.HelpBlock(),
		})
	case "start":
		err := bot.validateStartRequest(s.ChannelName)
		if err != nil {
			level.Info(bot.logger).Log("event", "start.rejected", "channel", s.ChannelName, "error", err)
			bot.sendJSON(w, http.StatusOK, &slack.Msg{
				ResponseType: slack.ResponseTypeEphemeral,
				Text:         "Icebreaker sessions can only be started in the #icebreakers channel.",
			})
			return
		}

		bot.startIcebreakerSession(command)
		w.WriteHeader(http.StatusOK)
	default:
		bot.sendJSON(w, http.StatusOK, &slack.Msg{
			ResponseType: slack.ResponseTypeEphemeral,
			Blocks:       message.HelpBlock(),
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

func (bot *BingsooBot) validateStartRequest(channelName string) error {
	if channelName != "icebreakers" {
		return fmt.Errorf("invalid channel")
	}
	return nil
}

func (bot *BingsooBot) startIcebreakerSession(command *command) error {
	level.Info(bot.logger).Log("event", "session.started")
	defer level.Info(bot.logger).Log("event", "session.finished")
	return nil
}
