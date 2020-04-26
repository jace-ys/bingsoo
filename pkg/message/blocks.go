package message

import (
	"fmt"

	"github.com/slack-go/slack"
)

func HelpBlock(channelID string) slack.Blocks {
	headerText := `
Hey there, I'm Bingsoo :wave::skin-tone-2:
I'm here to host icebreaker sessions to help you get to know your teammates better! :tada:
Use the following commands to interact with me.
`
	headerTextBlock := slack.NewTextBlockObject(slack.MarkdownType, headerText, false, false)
	headerSectionBlock := slack.NewSectionBlock(headerTextBlock, nil, nil)

	helpText := ":question: `/bingsoo help` displays useful information about me."
	helpTextBlock := slack.NewTextBlockObject(slack.MarkdownType, helpText, false, false)
	helpSectionBlock := slack.NewSectionBlock(helpTextBlock, nil, nil)

	startText := fmt.Sprintf(":shaved_ice: `/bingsoo start` starts an icebreaker session in the <#%s> channel.", channelID)
	startTextBlock := slack.NewTextBlockObject(slack.MarkdownType, startText, false, false)
	startSectionBlock := slack.NewSectionBlock(startTextBlock, nil, nil)

	return slack.Blocks{
		BlockSet: []slack.Block{
			headerSectionBlock,
			slack.NewDividerBlock(),
			helpSectionBlock,
			startSectionBlock,
		},
	}
}

func StartBlock(sessionID string) slack.Blocks {
	headerText := `
*:shaved_ice: It's time for some icebreakers! :shaved_ice:*
Suggest questions to ask your teammates or vote on your favourite ones.
People will be chosen at random to answer the selected question, and I will reveal their answers here at the end! :flushed:
`
	headerTextBlock := slack.NewTextBlockObject(slack.MarkdownType, headerText, false, false)
	headerSectionBlock := slack.NewSectionBlock(headerTextBlock, nil, nil)

	suggestButtonText := "Suggest a question"
	suggestButtonTextBlock := slack.NewTextBlockObject(slack.PlainTextType, suggestButtonText, false, false)
	suggestButtonBlockElement := slack.NewButtonBlockElement("", fmt.Sprintf("suggest-button/%s", sessionID), suggestButtonTextBlock)
	suggestButtonBlockElement.WithStyle(slack.StylePrimary)
	suggestButtonActionBlock := slack.NewActionBlock("", suggestButtonBlockElement)

	return slack.Blocks{
		BlockSet: []slack.Block{
			headerSectionBlock,
			slack.NewDividerBlock(),
			slack.NewDividerBlock(),
			suggestButtonActionBlock,
		},
	}
}
