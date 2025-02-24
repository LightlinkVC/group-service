package repository

import (
	"github.com/lightlink/group-service/internal/message/domain/entity"
	"github.com/lightlink/group-service/internal/message/domain/model"
)

type MessageRepositoryI interface {
	Create(messageEntity *entity.Message) (*model.Message, error)
	GetByGroupID(groupID uint) ([]entity.Message, error)
}
