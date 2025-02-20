package repository

import (
	"github.com/lightlink/group-service/internal/group/domain/entity"
	"github.com/lightlink/group-service/internal/group/domain/model"
)

type GroupRepositoryI interface {
	Create(groupEntity *entity.Group) (*model.Group, error)
}
