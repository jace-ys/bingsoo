package bingsoo

import (
	"net/http"
	"strings"

	"github.com/go-kit/kit/log/level"
	"github.com/slack-go/slack"
)

func (bot *BingsooBot) Commands(w http.ResponseWriter, r *http.Request) {
	level.Info(bot.logger).Log("event", "command.started")
	defer level.Info(bot.logger).Log("event", "command.finished")

	s, err := slack.SlashCommandParse(r)
	if err != nil {
		return
	}

	action, parameters := parseCommandText(s.Text)
	level.Info(bot.logger).Log("event", "command.parsed", "action", action, "parameters", parameters)

	switch action {
	case "help":
	case "start":
	default:
	}

	w.WriteHeader(http.StatusOK)
}

func parseCommandText(text string) (string, []string) {
	split := strings.Split(text, " ")
	return split[0], split[1:]
}
