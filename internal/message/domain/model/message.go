package model

type Message struct {
	ID      uint   `db:"id"`
	UserID  uint   `db:"user_id"`
	GroupID uint   `db:"group_id"`
	Content string `db:"content"`
}
