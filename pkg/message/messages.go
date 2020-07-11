package message

import (
	"fmt"
	"sort"

	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/interaction"
	"github.com/jace-ys/bingsoo/pkg/question"
)

func VoteMessage(sessionID string, questions question.QuestionSet) slack.MsgOption {
	var blocks []slack.Block

	headerText := `*:shaved_ice: It's time for some icebreakers! :shaved_ice:*
Suggest questions to ask your teammates or vote on your favourite ones.
People will be chosen at random to answer the selected question, and I will reveal their answers here at the end! :flushed:`
	headerTextBlock := slack.NewTextBlockObject(slack.MarkdownType, headerText, false, false)
	headerSectionBlock := slack.NewSectionBlock(headerTextBlock, nil, nil)
	blocks = append(blocks, headerSectionBlock, slack.NewDividerBlock())

	for _, question := range sortQuestions(questions) {
		voteButtonID := interaction.ActionVoteSubmit
		voteButtonTextBlock := slack.NewTextBlockObject(slack.PlainTextType, "Vote", false, false)
		voteButtonBlockElement := slack.NewButtonBlockElement(fmt.Sprintf("%s/%s", sessionID, voteButtonID), question, voteButtonTextBlock).WithStyle(slack.StylePrimary)

		questionTextBlock := slack.NewTextBlockObject(slack.MarkdownType, question, false, false)
		questionSectionBlock := slack.NewSectionBlock(questionTextBlock, nil, slack.NewAccessory(voteButtonBlockElement))

		voteCountText := fmt.Sprintf("%d vote(s)", len(questions[question]))
		voteCountTextBlock := slack.NewTextBlockObject(slack.PlainTextType, voteCountText, false, false)
		voteCountContextBlock := slack.NewContextBlock("", voteCountTextBlock)

		blocks = append(blocks, questionSectionBlock, voteCountContextBlock)
	}
	blocks = append(blocks, slack.NewDividerBlock())

	suggestButtonID := interaction.ActionSuggestionView
	suggestButtonTextBlock := slack.NewTextBlockObject(slack.PlainTextType, "Suggest a question", false, false)
	suggestButtonBlockElement := slack.NewButtonBlockElement(fmt.Sprintf("%s/%s", sessionID, suggestButtonID), "", suggestButtonTextBlock).WithStyle(slack.StylePrimary)
	suggestButtonActionBlock := slack.NewActionBlock(suggestButtonID, suggestButtonBlockElement)
	blocks = append(blocks, suggestButtonActionBlock)

	return slack.MsgOptionCompose(
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionText(headerText, false),
	)
}

func QuestionMessage(sessionID, channelID string) slack.MsgOption {
	var blocks []slack.Block

	headerText := fmt.Sprintf(`*An icebreaker session has been started in <#%s> and you have been selected! :shaved_ice:*
Answer the following question to participate.`, channelID)
	headerTextBlock := slack.NewTextBlockObject(slack.MarkdownType, headerText, false, false)
	headerSectionBlock := slack.NewSectionBlock(headerTextBlock, nil, nil)
	blocks = append(blocks, headerSectionBlock)

	answerButtonID := interaction.ActionQuestionView
	answerButtonTextBlock := slack.NewTextBlockObject(slack.PlainTextType, "Answer question", false, false)
	answerButtonBlockElement := slack.NewButtonBlockElement(fmt.Sprintf("%s/%s", sessionID, answerButtonID), "", answerButtonTextBlock).WithStyle(slack.StylePrimary)
	answerButtonActionBlock := slack.NewActionBlock(answerButtonID, answerButtonBlockElement)
	blocks = append(blocks, answerButtonActionBlock)

	return slack.MsgOptionCompose(
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionText(headerText, false),
	)
}

func ResultMessage(question string, responses map[string]string) slack.MsgOption {
	var blocks []slack.Block

	headerText := fmt.Sprintf(`:drum_with_drumsticks: *It's time! :drum_with_drumsticks: Revealing your teammates' responses to the icebreaker question:*
%s`, question)
	headerTextBlock := slack.NewTextBlockObject(slack.MarkdownType, headerText, false, false)
	headerSectionBlock := slack.NewSectionBlock(headerTextBlock, nil, nil)
	blocks = append(blocks, headerSectionBlock, slack.NewDividerBlock())

	empty := true
	for userID, response := range responses {
		if response != "" {
			responseText := fmt.Sprintf(`<@%s> answered: %s`, userID, response)
			responseTextBlock := slack.NewTextBlockObject(slack.MarkdownType, responseText, false, false)
			responseSectionBlock := slack.NewSectionBlock(responseTextBlock, nil, nil)
			blocks = append(blocks, responseSectionBlock)

			empty = false
		}
	}

	if empty {
		responseText := `Hmm... it seems like no one responded.
Did someone forget to tell me today was a holiday? :see_no_evil:`
		responseTextBlock := slack.NewTextBlockObject(slack.MarkdownType, responseText, false, false)
		responseSectionBlock := slack.NewSectionBlock(responseTextBlock, nil, nil)
		blocks = append(blocks, responseSectionBlock)
	}

	return slack.MsgOptionCompose(
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionText(headerText, false),
	)
}

func ErrorMessage() slack.MsgOption {
	var blocks []slack.Block

	headerText := `Seems like an unexpected error has occurred :disappointed:
The icebreaker session has been ended, please try again later.`
	headerTextBlock := slack.NewTextBlockObject(slack.MarkdownType, headerText, false, false)
	headerSectionBlock := slack.NewSectionBlock(headerTextBlock, nil, nil)
	blocks = append(blocks, headerSectionBlock)

	return slack.MsgOptionCompose(
		slack.MsgOptionBlocks(blocks...),
		slack.MsgOptionText(headerText, false),
	)
}

func sortQuestions(questions question.QuestionSet) []string {
	type pair struct {
		key   string
		value int
	}

	var pairs []pair
	for question, users := range questions {
		pairs = append(pairs, pair{question, len(users)})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value > pairs[j].value
	})

	sorted := make([]string, len(questions))
	for idx, pair := range pairs {
		sorted[idx] = pair.key
	}

	return sorted
}
