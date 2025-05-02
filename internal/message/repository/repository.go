package repository

import (
	"github.com/lightlink/group-service/internal/message/domain/entity"
)

type MessageRepositoryI interface {
	Create(messageEntity *entity.Message) (*entity.Message, error)
	GetByGroupID(groupID uint) ([]entity.Message, error)
	UpdateStatus(messageID uint, statusName string) error
}
