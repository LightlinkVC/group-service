package usecase

import (
	"github.com/lightlink/group-service/internal/group/domain/entity"
	"github.com/lightlink/group-service/internal/group/repository"
)

type GroupUsecaseI interface {
	Create(groupEntity *entity.Group, groupMembers []entity.GroupMember) error
	GetPersonalGroupID(user1ID uint, user2ID uint) (uint, error)
}

type GroupUsecase struct {
	groupRepo repository.GroupRepositoryI
}

func NewGroupUsecase(groupRepository repository.GroupRepositoryI) *GroupUsecase {
	return &GroupUsecase{
		groupRepo: groupRepository,
	}
}

func (uc *GroupUsecase) Create(groupEntity *entity.Group, groupMembers []entity.GroupMember) error {
	_, err := uc.groupRepo.Create(groupEntity, groupMembers)
	if err != nil {
		return err
	}

	return nil
}

func (uc *GroupUsecase) GetPersonalGroupID(user1ID uint, user2ID uint) (uint, error) {
	groupID, err := uc.groupRepo.GetPersonalGroupID(user1ID, user2ID)
	if err != nil {
		return 0, err
	}

	return groupID, nil
}
