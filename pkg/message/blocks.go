package message

import (
	"fmt"

	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/question"
)

func HelpBlock(channelID string) slack.Blocks {
	var blocks []slack.Block

	headerText := `
Hey there, I'm Bingsoo :wave::skin-tone-2:
I'm here to host icebreaker sessions to help you get to know your teammates better! :tada:
Use the following commands to interact with me.
`
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

	return slack.Blocks{
		BlockSet: blocks,
	}
}

func VoteBlock(sessionID string, questions []*question.Question) slack.Blocks {
	var blocks []slack.Block

	headerText := `
*:shaved_ice: It's time for some icebreakers! :shaved_ice:*
Suggest questions to ask your teammates or vote on your favourite ones.
People will be chosen at random to answer the selected question, and I will reveal their answers here at the end! :flushed:
`
	headerTextBlock := slack.NewTextBlockObject(slack.MarkdownType, headerText, false, false)
	headerSectionBlock := slack.NewSectionBlock(headerTextBlock, nil, nil)
	blocks = append(blocks, headerSectionBlock, slack.NewDividerBlock())

	for _, question := range questions {
		voteButtonTextBlock := slack.NewTextBlockObject(slack.PlainTextType, "Vote", false, false)
		voteButtonBlockElement := slack.NewButtonBlockElement(fmt.Sprintf("%s/vote/%d", sessionID, question.ID), question.Value, voteButtonTextBlock)
		voteButtonBlockElement.WithStyle(slack.StylePrimary)

		questionTextBlock := slack.NewTextBlockObject(slack.MarkdownType, question.Value, false, false)
		questionSectionBlock := slack.NewSectionBlock(questionTextBlock, nil, slack.NewAccessory(voteButtonBlockElement))
		blocks = append(blocks, questionSectionBlock)
	}
	blocks = append(blocks, slack.NewDividerBlock())

	suggestButtonTextBlock := slack.NewTextBlockObject(slack.PlainTextType, "Suggest a question", false, false)
	suggestButtonBlockElement := slack.NewButtonBlockElement(fmt.Sprintf("%s/suggest", sessionID), "suggest", suggestButtonTextBlock)
	suggestButtonBlockElement.WithStyle(slack.StylePrimary)
	suggestButtonActionBlock := slack.NewActionBlock("", suggestButtonBlockElement)
	blocks = append(blocks, suggestButtonActionBlock)

	return slack.Blocks{
		BlockSet: blocks,
	}
}

func QuestionBlock(channelID string, question *question.Question) slack.Blocks {
	var blocks []slack.Block

	headerText := `
*An icebreaker session has been started in <#%s> and you have been selected!*
*Here's your question:* %s
`
	headerTextBlock := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf(headerText, channelID, question.Value), false, false)
	headerSectionBlock := slack.NewSectionBlock(headerTextBlock, nil, nil)
	blocks = append(blocks, headerSectionBlock)

	return slack.Blocks{
		BlockSet: blocks,
	}
}

func ResultBlock(question *question.Question) slack.Blocks {
	var blocks []slack.Block

	headerText := `
:drum_with_drumsticks: *It's time! :drum_with_drumsticks: Revealing your teammates' responses to the icebreaker question:*
%s
`
	headerTextBlock := slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf(headerText, question.Value), false, false)
	headerSectionBlock := slack.NewSectionBlock(headerTextBlock, nil, nil)
	blocks = append(blocks, headerSectionBlock, slack.NewDividerBlock())

	return slack.Blocks{
		BlockSet: blocks,
	}
}
