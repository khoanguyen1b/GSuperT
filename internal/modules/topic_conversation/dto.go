package topic_conversation

type GenerateRequest struct {
	Topic string `json:"topic"`
	Turns int    `json:"turns,omitempty"`
}

type GenerateResponse struct {
	Valid             bool           `json:"valid"`
	Topic             string         `json:"topic,omitempty"`
	ValidationMessage string         `json:"validation_message"`
	TurnCount         int            `json:"turn_count,omitempty"`
	Turns             []DialogueTurn `json:"turns,omitempty"`
}

type DialogueTurn struct {
	Turn    int    `json:"turn"`
	Speaker string `json:"speaker"`
	TextEN  string `json:"text_en"`
	TextVI  string `json:"text_vi"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}
