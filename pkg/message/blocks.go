package message

import (
	"fmt"
	"sort"

	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/interaction"
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

	return slack.Blocks{blocks}
}

func VoteBlock(sessionID string, questions question.QuestionSet) slack.Blocks {
	var blocks []slack.Block

	headerText := `
*:shaved_ice: It's time for some icebreakers! :shaved_ice:*
Suggest questions to ask your teammates or vote on your favourite ones.
People will be chosen at random to answer the selected question, and I will reveal their answers here at the end! :flushed:`
	headerTextBlock := slack.NewTextBlockObject(slack.MarkdownType, headerText, false, false)
	headerSectionBlock := slack.NewSectionBlock(headerTextBlock, nil, nil)
	blocks = append(blocks, headerSectionBlock, slack.NewDividerBlock())

	for _, question := range sortQuestions(questions) {
		voteButtonID := interaction.ActionVoteSubmit
		voteButtonTextBlock := slack.NewTextBlockObject(slack.PlainTextType, "Vote", false, false)
		voteButtonBlockElement := slack.NewButtonBlockElement(fmt.Sprintf("%s/%s", sessionID, voteButtonID), question, voteButtonTextBlock)
		voteButtonBlockElement.WithStyle(slack.StylePrimary)

		questionTextBlock := slack.NewTextBlockObject(slack.MarkdownType, question, false, false)
		questionSectionBlock := slack.NewSectionBlock(questionTextBlock, nil, slack.NewAccessory(voteButtonBlockElement))
		blocks = append(blocks, questionSectionBlock)
	}
	blocks = append(blocks, slack.NewDividerBlock())

	suggestButtonBlockID := interaction.ActionSuggestionView
	suggestButtonTextBlock := slack.NewTextBlockObject(slack.PlainTextType, "Suggest a question", false, false)
	suggestButtonBlockElement := slack.NewButtonBlockElement(fmt.Sprintf("%s/%s", sessionID, suggestButtonBlockID), "", suggestButtonTextBlock)
	suggestButtonBlockElement.WithStyle(slack.StylePrimary)
	suggestButtonActionBlock := slack.NewActionBlock(suggestButtonBlockID, suggestButtonBlockElement)
	blocks = append(blocks, suggestButtonActionBlock)

	return slack.Blocks{blocks}
}

func QuestionBlock(sessionID, channelID string) slack.Blocks {
	var blocks []slack.Block

	headerText := fmt.Sprintf(`
*An icebreaker session has been started in <#%s> and you have been selected! :shaved_ice:*
Answer the following question to participate.`, channelID)
	headerTextBlock := slack.NewTextBlockObject(slack.MarkdownType, headerText, false, false)
	headerSectionBlock := slack.NewSectionBlock(headerTextBlock, nil, nil)
	blocks = append(blocks, headerSectionBlock)

	answerButtonBlockID := interaction.ActionQuestionView
	answerButtonTextBlock := slack.NewTextBlockObject(slack.PlainTextType, "Answer question", false, false)
	answerButtonBlockElement := slack.NewButtonBlockElement(fmt.Sprintf("%s/%s", sessionID, answerButtonBlockID), "", answerButtonTextBlock)
	answerButtonBlockElement.WithStyle(slack.StylePrimary)
	answerButtonActionBlock := slack.NewActionBlock(answerButtonBlockID, answerButtonBlockElement)
	blocks = append(blocks, answerButtonActionBlock)

	return slack.Blocks{blocks}
}

func ResultBlock(question string, responses map[string]string) slack.Blocks {
	var blocks []slack.Block

	headerText := fmt.Sprintf(`
:drum_with_drumsticks: *It's time! :drum_with_drumsticks: Revealing your teammates' responses to the icebreaker question:*
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
		responseText := `
Hmm... it seems like no one responded.
Did someone forget to tell me today was a holiday? :see_no_evil:`
		responseTextBlock := slack.NewTextBlockObject(slack.MarkdownType, responseText, false, false)
		responseSectionBlock := slack.NewSectionBlock(responseTextBlock, nil, nil)
		blocks = append(blocks, responseSectionBlock)
	}

	return slack.Blocks{blocks}
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
