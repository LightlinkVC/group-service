package dto

type CreateMessageRequest struct {
	UserID  uint   `json:"user_id"`
	GroupID uint   `json:"group_id"`
	Content string `json:"content"`
}
