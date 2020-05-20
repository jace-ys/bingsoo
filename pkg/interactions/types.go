package interactions

type InteractionType string

const (
	Action   InteractionType = "Action"
	Response InteractionType = "Response"
)

const (
	ActionQuestionView string = "question/view"
)

const (
	ResponseAnswerSubmit string = "answer/submit"
)
