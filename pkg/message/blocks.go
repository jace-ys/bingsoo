package message

import (
	"fmt"

	"github.com/slack-go/slack"
)

func HelpBlock(channelID string) slack.Blocks {
	var blocks []slack.Block

	headerText := `Hey there, I'm Bingsoo :wave::skin-tone-2:
I'm here to host icebreaker sessions to help you get to know your teammates better! :tada:
Use the following commands to interact with me.`
	headerTextBlock := slack.NewTextBlockObject(slack.MarkdownType, headerText, false, false)
	headerSectionBlock := slack.NewSectionBlock(headerTextBlock, nil, nil)
	blocks = append(blocks, headerSectionBlock, slack.NewDividerBlock())

	helpText := ":question: `/bingsoo help` displays useful information about me."
	helpTextBlock := slack.NewTextBlockObject(slack.MarkdownType, helpText, false, false)
	helpSectionBlock := slack.NewSectionBlock(helpTextBlock, nil, nil)
	blocks = append(blocks, helpSectionBlock)

	startText := fmt.Sprintf(":shaved_ice: `/bingsoo start` starts an icebreaker session in the <#%s> channel.", channelID)
	startTextBlock := slack.NewTextBlockObject(slack.MarkdownType, startText, false, false)
	startSectionBlock := slack.NewSectionBlock(startTextBlock, nil, nil)
	blocks = append(blocks, startSectionBlock)

	return slack.Blocks{BlockSet: blocks}
}
