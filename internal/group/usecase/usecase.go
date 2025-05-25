package usecase

import (
	"fmt"
	"strconv"

	"github.com/lightlink/group-service/internal/group/domain/entity"
	groupRepo "github.com/lightlink/group-service/internal/group/repository"
	notificationDTO "github.com/lightlink/group-service/internal/notification/domain/dto"
	notificationRepo "github.com/lightlink/group-service/internal/notification/repository"
)

type GroupUsecaseI interface {
	Create(groupEntity *entity.Group, groupMembers []entity.GroupMember) error
	GetGroupsByUserID(userID uint) ([]entity.Group, error)
	GetPersonalGroupID(user1ID uint, user2ID uint) (uint, error)
	StartCall(initiatorIDString, groupIDString string) error
}

type GroupUsecase struct {
	groupRepo        groupRepo.GroupRepositoryI
	notificationRepo notificationRepo.NotificationRepositoryI
}

func NewGroupUsecase(
	groupRepository groupRepo.GroupRepositoryI,
	notificationRepo notificationRepo.NotificationRepositoryI,
) *GroupUsecase {
	return &GroupUsecase{
		groupRepo:        groupRepository,
		notificationRepo: notificationRepo,
	}
}

func (uc *GroupUsecase) StartCall(initiatorIDString, groupIDString string) error {
	groupID, err := strconv.ParseUint(groupIDString, 10, 64)
	if err != nil {
		return err
	}

	initiatorID, err := strconv.ParseUint(initiatorIDString, 10, 64)
	if err != nil {
		return err
	}

	err = uc.sendIncomingCallNotification(uint(initiatorID), uint(groupID))
	if err != nil {
		return err
	}

	return nil
}

func (uc *GroupUsecase) sendIncomingCallNotification(initiatorID, groupID uint) error {
	memberIDs, err := uc.groupRepo.GetMemberIDsByGroupID(uint(groupID))
	if err != nil {
		return err
	}

	for _, memberID := range memberIDs {
		if memberID == initiatorID {
			continue
		}

		notifErr := uc.notificationRepo.Send(notificationDTO.RawNotification{
			Type: "incomingCall",
			Payload: map[string]interface{}{
				"from_user_id": strconv.FormatUint(uint64(initiatorID), 10),
				"to_user_id":   strconv.FormatUint(uint64(memberID), 10),
				"room_id":      strconv.FormatUint(uint64(groupID), 10),
			},
		})
		if notifErr != nil {
			fmt.Println("Error sending incomingCall notif in kafka")
		}
	}

	return nil
}

func (uc *GroupUsecase) Create(groupEntity *entity.Group, groupMembers []entity.GroupMember) error {
	_, err := uc.groupRepo.Create(groupEntity, groupMembers)
	if err != nil {
		return err
	}

	return nil
}

func (uc *GroupUsecase) GetGroupsByUserID(userID uint) ([]entity.Group, error) {
	groupModels, err := uc.groupRepo.GetGroupsByUserID(userID)
	if err != nil {
		return nil, err
	}

	groupEntities := []entity.Group{}
	for _, groupModel := range groupModels {
		groupEntity := entity.Group{
			ID:        groupModel.ID,
			Name:      groupModel.Name,
			CreatorID: groupModel.CreatorID,
			TypeName:  "group", // TODO
		}

		groupEntities = append(groupEntities, groupEntity)
	}

	return groupEntities, nil
}

func (uc *GroupUsecase) GetPersonalGroupID(user1ID uint, user2ID uint) (uint, error) {
	groupID, err := uc.groupRepo.GetPersonalGroupID(user1ID, user2ID)
	if err != nil {
		return 0, err
	}

	return groupID, nil
}
