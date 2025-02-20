package usecase

import (
	"github.com/lightlink/group-service/internal/group/domain/dto"
	"github.com/lightlink/group-service/internal/group/repository"
)

type GroupUsecaseI interface {
	Create(createRequest *dto.CreateGroupRequest) error
}

type GroupUsecase struct {
	groupRepo repository.GroupRepositoryI
}

func NewGroupUsecase(groupRepository repository.GroupRepositoryI) *GroupUsecase {
	return &GroupUsecase{
		groupRepo: groupRepository,
	}
}

func (uc *GroupUsecase) Create(createRequest *dto.CreateGroupRequest) error {
	groupEntity := dto.CreateGroupRequestToEntity(createRequest)
	_, err := uc.groupRepo.Create(groupEntity)
	if err != nil {
		return err
	}

	return nil
}
