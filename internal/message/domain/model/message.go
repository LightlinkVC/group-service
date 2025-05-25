package model

import "time"

type Message struct {
	ID        uint      `db:"id"`
	UserID    uint      `db:"user_id"`
	GroupID   uint      `db:"group_id"`
	StatusID  uint      `db:"status_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `json:"created_at"`
}
