package usecase

import (
	"log"
	"strconv"

	groupRepo "github.com/lightlink/group-service/internal/group/repository"
	messageDTO "github.com/lightlink/group-service/internal/message/domain/dto"
	"github.com/lightlink/group-service/internal/message/domain/entity"
	messageRepo "github.com/lightlink/group-service/internal/message/repository"
	notificationDTO "github.com/lightlink/group-service/internal/notification/domain/dto"
	notificationRepo "github.com/lightlink/group-service/internal/notification/repository"
)

type MessageUsecaseI interface {
	Create(createRequest *messageDTO.CreateMessageRequest) (*entity.Message, error)
	GetByGroupID(groupID uint) ([]entity.Message, error)
}

type MessageUsecase struct {
	messageRepo      messageRepo.MessageRepositoryI
	groupRepo        groupRepo.GroupRepositoryI
	notificationRepo notificationRepo.NotificationRepositoryI
}

func NewMessageUsecase(
	messageRepo messageRepo.MessageRepositoryI,
	notificationRepo notificationRepo.NotificationRepositoryI,
	groupRepo groupRepo.GroupRepositoryI,
) *MessageUsecase {
	return &MessageUsecase{
		messageRepo:      messageRepo,
		notificationRepo: notificationRepo,
		groupRepo:        groupRepo,
	}
}

func (uc *MessageUsecase) sendIncomingMessageNotification(senderID, roomID uint, content string) {
	receiverIDs, err := uc.groupRepo.GetMemberIDsByGroupID(roomID)
	if err != nil {
		log.Printf("ERR: An error occured due selecting group members: %v\n", err)
	}

	for _, receiverID := range receiverIDs {
		if receiverID == senderID {
			continue
		}

		uc.notificationRepo.Send(notificationDTO.RawNotification{
			Type: "incomingMessage",
			Payload: map[string]interface{}{
				"from_user_id": strconv.FormatUint(uint64(senderID), 10),
				"to_user_id":   strconv.FormatUint(uint64(receiverID), 10),
				"room_id":      strconv.FormatUint(uint64(roomID), 10),
				"content":      content,
			},
		})
	}
}

func (uc *MessageUsecase) Create(createRequest *messageDTO.CreateMessageRequest) (*entity.Message, error) {
	messageEntity := entity.Message{
		UserID:  createRequest.UserID,
		GroupID: createRequest.GroupID,
		Content: createRequest.Content,
	}

	createdMessageModel, err := uc.messageRepo.Create(&messageEntity)
	if err != nil {
		return nil, err
	}

	go uc.sendIncomingMessageNotification(
		messageEntity.UserID,
		messageEntity.GroupID,
		messageEntity.Content,
	)

	createdMessageEntity := messageDTO.MessageModelToEntity(createdMessageModel)

	return createdMessageEntity, nil
}

func (uc *MessageUsecase) GetByGroupID(groupID uint) ([]entity.Message, error) {
	messages, err := uc.messageRepo.GetByGroupID(groupID)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
