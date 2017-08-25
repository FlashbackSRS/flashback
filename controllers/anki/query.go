package anki

type basicQuery struct {
	Submit       string
	TypedAnswers map[string]string
}

const typePrefix = "type:"
