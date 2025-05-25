package repository

import (
	"github.com/lightlink/group-service/internal/group/domain/entity"
	"github.com/lightlink/group-service/internal/group/domain/model"
)

type GroupRepositoryI interface {
	Create(groupEntity *entity.Group, groupMembers []entity.GroupMember) (*model.Group, error)
	GetPersonalGroupID(user1ID uint, user2ID uint) (uint, error)
	GetMemberIDsByGroupID(groupID uint) ([]uint, error)
}
