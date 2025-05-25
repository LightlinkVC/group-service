package dto

import "mime/multipart"

type CreateMessageRequest struct {
	UserID  uint                    `json:"user_id"`
	GroupID uint                    `json:"group_id"`
	Content string                  `json:"content"`
	Files   []*multipart.FileHeader `form:"files"`
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

type FileInfo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type"`
	Size int64  `json:"size"`
}

type IncomingMessagePayload struct {
	ID      uint       `json:"id"`
	UserID  uint       `json:"user_id"`
	GroupID uint       `json:"group_id"`
	Status  string     `json:"status"`
	Content string     `json:"content"`
	Files   []FileInfo `json:"files"`
}

type HateSpeechStatusAckPayload struct {
	MessageID uint `json:"message_id"`
}
