package entity

type Message struct {
	UserID  uint   `json:"user_id"`
	GroupID uint   `json:"group_id"`
	Content string `json:"content"`
}
