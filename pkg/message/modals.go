package message

import (
	"fmt"

	"github.com/slack-go/slack"

	"github.com/jace-ys/bingsoo/pkg/interactions"
	"github.com/jace-ys/bingsoo/pkg/question"
)

func AnswerModal(sessionID string, question *question.Question) slack.ModalViewRequest {
	var blocks []slack.Block

	answerInputBlockID := interactions.ResponseAnswerSubmit
	answerLabelTextBlock := slack.NewTextBlockObject(slack.PlainTextType, question.Value, false, false)
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
