package message

import (
	"fmt"

	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/interaction"
)

func SuggestionModal(sessionID string) slack.ModalViewRequest {
	var blocks []slack.Block

	suggestionInputBlockID := interaction.ResponseSuggestionSubmit
	suggestionLabelTextBlock := slack.NewTextBlockObject(slack.PlainTextType, "Suggest a question", false, false)
	suggestionPlaceholderTextBlock := slack.NewTextBlockObject(slack.PlainTextType, "Enter your question here", false, false)
	suggestionInputBlockElement := slack.NewPlainTextInputBlockElement(suggestionPlaceholderTextBlock, fmt.Sprintf("%s/%s", sessionID, suggestionInputBlockID))
	suggestionInputBlock := slack.NewInputBlock(suggestionInputBlockID, suggestionLabelTextBlock, suggestionInputBlockElement)
	blocks = append(blocks, suggestionInputBlock)

	return slack.ModalViewRequest{
		Type: slack.VTModal,
		Title: &slack.TextBlockObject{
			Type:  slack.PlainTextType,
			Text:  "Suggest a question!",
			Emoji: true,
		},
		Blocks: slack.Blocks{BlockSet: blocks},
		Submit: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Submit",
		},
		Close: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Cancel",
		},
	}
}

func AnswerModal(sessionID string, question string) slack.ModalViewRequest {
	var blocks []slack.Block

	answerInputBlockID := interaction.ResponseAnswerSubmit
	answerLabelTextBlock := slack.NewTextBlockObject(slack.PlainTextType, question, false, false)
	answerPlaceholderTextBlock := slack.NewTextBlockObject(slack.PlainTextType, "Enter your answer here", false, false)
	answerInputBlockElement := slack.NewPlainTextInputBlockElement(answerPlaceholderTextBlock, fmt.Sprintf("%s/%s", sessionID, answerInputBlockID))
	answerInputBlock := slack.NewInputBlock(answerInputBlockID, answerLabelTextBlock, answerInputBlockElement)
	blocks = append(blocks, answerInputBlock)

	return slack.ModalViewRequest{
		Type: slack.VTModal,
		Title: &slack.TextBlockObject{
			Type:  slack.PlainTextType,
			Text:  "Here's your question!",
			Emoji: true,
		},
		Blocks: slack.Blocks{BlockSet: blocks},
		Submit: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Submit",
		},
		Close: &slack.TextBlockObject{
			Type: slack.PlainTextType,
			Text: "Cancel",
		},
	}
}
