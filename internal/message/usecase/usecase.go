package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/lightlink/group-service/infrastructure/ws"
	groupRepo "github.com/lightlink/group-service/internal/group/repository"
	messageDTO "github.com/lightlink/group-service/internal/message/domain/dto"
	"github.com/lightlink/group-service/internal/message/domain/entity"
	messageRepo "github.com/lightlink/group-service/internal/message/repository"
	notificationDTO "github.com/lightlink/group-service/internal/notification/domain/dto"
	notificationRepo "github.com/lightlink/group-service/internal/notification/repository"
)

const (
	HATE_MESSAGE_STATUS    = "hate"
	NEUTRAL_MESSAGE_STATUS = "neutral"
)

type MessageUsecaseI interface {
	Create(createRequest *messageDTO.CreateMessageRequest) (*entity.Message, error)
	GetByGroupID(groupID uint) ([]entity.Message, error)
	UpdateHateSpeechLabel(hateSpeechResponse messageDTO.MessageHateSpeechResponse)
}

type MessageUsecase struct {
	messageRepo           messageRepo.MessageRepositoryI
	groupRepo             groupRepo.GroupRepositoryI
	notificationRepo      notificationRepo.NotificationRepositoryI
	messageHateSpeechRepo messageRepo.MessageHateSpeechRepositoryI
	messagingServer       ws.MessagingServer
}

func NewMessageUsecase(
	messageRepo messageRepo.MessageRepositoryI,
	notificationRepo notificationRepo.NotificationRepositoryI,
	groupRepo groupRepo.GroupRepositoryI,
	messageHateSpeechRepo messageRepo.MessageHateSpeechRepositoryI,
	messagingServer ws.MessagingServer,
) *MessageUsecase {
	return &MessageUsecase{
		messageRepo:           messageRepo,
		notificationRepo:      notificationRepo,
		groupRepo:             groupRepo,
		messageHateSpeechRepo: messageHateSpeechRepo,
		messagingServer:       messagingServer,
	}
}

func (uc *MessageUsecase) initiateHateSpeechCheck(messageID, groupID uint, content string) {
	hateSpeechRequest := messageDTO.MessageHateSpeechRequest{
		ID:      messageID,
		GroupID: groupID,
		Content: content,
	}

	err := uc.messageHateSpeechRepo.Send(hateSpeechRequest)
	if err != nil {
		log.Printf("ERR: An error occured sending message in hate-speech-service: %v\n", err)
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

	createdMessageEntity, err := uc.messageRepo.Create(&messageEntity)
	if err != nil {
		return nil, err
	}

	messagePayload, err := json.Marshal(createdMessageEntity)
	if err != nil {
		return nil, err
	}

	go uc.sendIncomingMessageNotification(
		createdMessageEntity.UserID,
		createdMessageEntity.GroupID,
		createdMessageEntity.Content,
	)

	go uc.initiateHateSpeechCheck(
		createdMessageEntity.ID,
		createdMessageEntity.GroupID,
		createdMessageEntity.Content,
	)

	uc.messagingServer.PublishToGroup(createdMessageEntity.GroupID, messagePayload)

	return createdMessageEntity, nil
}

func (uc *MessageUsecase) GetByGroupID(groupID uint) ([]entity.Message, error) {
	messages, err := uc.messageRepo.GetByGroupID(groupID)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (uc *MessageUsecase) UpdateHateSpeechLabel(hateSpeechResponse messageDTO.MessageHateSpeechResponse) {
	var newStatus string
	if hateSpeechResponse.IsHateSpeech {
		newStatus = HATE_MESSAGE_STATUS
	} else {
		newStatus = NEUTRAL_MESSAGE_STATUS
	}

	err := uc.messageRepo.UpdateStatus(hateSpeechResponse.ID, newStatus)
	if err != nil {
		fmt.Printf("Failed to update status for message %d: %v\n", hateSpeechResponse.ID, err)
		return
	}

	// uc.messagingServer.PublishToGroup()
}
