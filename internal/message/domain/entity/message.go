package entity

import "time"

type Message struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	GroupID   uint      `json:"group_id"`
	Status    string    `json:"status"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
