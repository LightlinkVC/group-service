package entity

import (
	"time"
)

type Message struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	GroupID   uint      `json:"group_id"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	Files     []File    `json:"files"`
}

type File struct {
	ID           uint   `json:"id"`
	ObjectName   string `json:"-"`
	OriginalName string `json:"name"`
	ContentType  string `json:"type"`
	Size         int64  `json:"size"`
	URL          string `json:"url"`
}
