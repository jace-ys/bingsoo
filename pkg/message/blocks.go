package message

import (
	"fmt"

	"github.com/slack-go/slack"
)

func HelpBlock(channelID string) slack.Blocks {
	headerText := "Hey there, I'm Bingsoo :wave::skin-tone-2: I'm here to host interactive icebreaker sessions to help you get to know your teammates better! :tada:"
	helpText := ":question: `/bingsoo help` displays useful information on using Bingsoo."
	startText := fmt.Sprintf(":shaved_ice: `/bingsoo start` starts an icebreaker session in the <#%s> channel.", channelID)

	headerTextBlock := slack.NewTextBlockObject(slack.MarkdownType, headerText, false, false)
	headerSectionBlock := slack.NewSectionBlock(headerTextBlock, nil, nil)

	helpTextBlock := slack.NewTextBlockObject(slack.MarkdownType, helpText, false, false)
	helpSectionBlock := slack.NewSectionBlock(helpTextBlock, nil, nil)

	startTextBlock := slack.NewTextBlockObject(slack.MarkdownType, startText, false, false)
	startSectionBlock := slack.NewSectionBlock(startTextBlock, nil, nil)

	return slack.Blocks{
		BlockSet: []slack.Block{
			headerSectionBlock,
			helpSectionBlock,
			startSectionBlock,
		},
	}
}
