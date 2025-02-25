package entity

type Message struct {
	ID      uint   `json:"id"`
	UserID  uint   `json:"user_id"`
	GroupID uint   `json:"group_id"`
	Content string `json:"content"`
}
