package interaction

type Action = string

const (
	ActionVoteSubmit     Action = "vote/submit"
	ActionSuggestionView Action = "suggest/view"
	ActionQuestionView   Action = "question/view"
)

type Response = string

const (
	ResponseSuggestionSubmit Response = "suggest/submit"
	ResponseAnswerSubmit     Response = "answer/submit"
)
