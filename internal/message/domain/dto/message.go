package dto

type CreateMessageRequest struct {
	UserID  uint   `json:"user_id"`
	GroupID uint   `json:"group_id"`
	Content string `json:"content"`
}

type MessageHateSpeechRequest struct {
	ID      uint   `json:"id"`
	GroupID uint   `json:"group_id"`
	Content string `json:"content"`
}

type MessageHateSpeechResponse struct {
	ID           uint `json:"id"`
	GroupID      uint `json:"group_id"`
	IsHateSpeech bool `json:"is_hate_speech"`
}

type MessageSignal struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type IncomingMessagePayload struct {
	ID      uint   `json:"id"`
	UserID  uint   `json:"user_id"`
	GroupID uint   `json:"group_id"`
	Status  string `json:"status"`
	Content string `json:"content"`
}

type HateSpeechStatusAckPayload struct {
	MessageID uint `json:"message_id"`
}
