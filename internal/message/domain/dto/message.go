package dto

import (
	"github.com/lightlink/group-service/internal/message/domain/entity"
	"github.com/lightlink/group-service/internal/message/domain/model"
)

type CreateMessageRequest struct {
	UserID  uint   `json:"user_id"`
	GroupID uint   `json:"group_id"`
	Content string `json:"content"`
}

func MessageModelToEntity(messageModel *model.Message) *entity.Message {
	return &entity.Message{
		ID:      messageModel.ID,
		UserID:  messageModel.UserID,
		GroupID: messageModel.GroupID,
		Content: messageModel.Content,
	}
}
