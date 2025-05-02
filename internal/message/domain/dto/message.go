package dto

type CreateMessageRequest struct {
	UserID  uint   `json:"user_id"`
	GroupID uint   `json:"group_id"`
	Content string `json:"content"`
}

type MessageHateSpeechRequest struct {
	ID      uint   `json:"id"`
	Content string `json:"content"`
}

type MessageHateSpeechResponse struct {
	ID           uint `json:"id"`
	IsHateSpeech bool `json:"is_hate_speech"`
}
