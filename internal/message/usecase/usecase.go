package usecase

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/lightlink/group-service/infrastructure/ws"
	fileRepo "github.com/lightlink/group-service/internal/file/repository"
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
	fileRepo              fileRepo.FileRepositoryI
	notificationRepo      notificationRepo.NotificationRepositoryI
	messageHateSpeechRepo messageRepo.MessageHateSpeechRepositoryI
	messagingServer       ws.MessagingServer
}

func NewMessageUsecase(
	messageRepo messageRepo.MessageRepositoryI,
	notificationRepo notificationRepo.NotificationRepositoryI,
	groupRepo groupRepo.GroupRepositoryI,
	fileRepo fileRepo.FileRepositoryI,
	messageHateSpeechRepo messageRepo.MessageHateSpeechRepositoryI,
	messagingServer ws.MessagingServer,
) *MessageUsecase {
	return &MessageUsecase{
		messageRepo:           messageRepo,
		notificationRepo:      notificationRepo,
		groupRepo:             groupRepo,
		fileRepo:              fileRepo,
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
		Files:   make([]entity.File, 0, len(createRequest.Files)),
	}

	for _, fileHeader := range createRequest.Files {
		file, err := fileHeader.Open()
		if err != nil {
			log.Printf("Failed to open file: %v", err)
			continue
		}
		defer file.Close()

		objectName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), fileHeader.Filename)

		err = uc.fileRepo.UploadObject(
			objectName,
			file,
			fileHeader.Size,
			fileHeader.Header.Get("Content-Type"),
		)
		if err != nil {
			log.Printf("Failed to upload file: %v", err)
			continue
		}

		url, err := uc.fileRepo.GetPresignedURL(objectName, 24*time.Hour)
		if err != nil {
			log.Printf("Failed to generate URL for file: %v", err)
			continue
		}

		messageEntity.Files = append(messageEntity.Files, entity.File{
			ObjectName:   objectName,
			OriginalName: fileHeader.Filename,
			ContentType:  fileHeader.Header.Get("Content-Type"),
			Size:         fileHeader.Size,
			URL:          url,
		})
	}

	createdMessageEntity, err := uc.messageRepo.Create(&messageEntity)
	if err != nil {
		return nil, err
	}

	filesWithURLs := make([]messageDTO.FileInfo, 0, len(createdMessageEntity.Files))
	for _, file := range createdMessageEntity.Files {
		url, err := uc.fileRepo.GetPresignedURL(file.ObjectName, 24*time.Hour)
		if err == nil {
			filesWithURLs = append(filesWithURLs, messageDTO.FileInfo{
				Name: file.OriginalName,
				URL:  url,
				Type: file.ContentType,
				Size: file.Size,
			})
		}
	}

	messagePayload := messageDTO.IncomingMessagePayload{
		ID:      createdMessageEntity.ID,
		UserID:  createdMessageEntity.UserID,
		GroupID: createdMessageEntity.GroupID,
		Status:  createdMessageEntity.Status,
		Content: createRequest.Content,
		Files:   filesWithURLs,
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

	uc.messagingServer.PublishToGroup(
		createdMessageEntity.GroupID,
		messageDTO.MessageSignal{
			Type:    "newMessage",
			Payload: messagePayload,
		},
	)

	return createdMessageEntity, nil
}

func (uc *MessageUsecase) GetByGroupID(groupID uint) ([]entity.Message, error) {
	messages, err := uc.messageRepo.GetByGroupID(groupID)
	if err != nil {
		return nil, err
	}

	for i := range messages {
		for j := range messages[i].Files {
			url, err := uc.fileRepo.GetPresignedURL(
				messages[i].Files[j].ObjectName,
				24*time.Hour,
			)
			if err == nil {
				messages[i].Files[j].URL = url
			}
		}
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

	if newStatus != HATE_MESSAGE_STATUS {
		return
	}

	hateMessagecknowledgementPayload := messageDTO.HateSpeechStatusAckPayload{
		MessageID: hateSpeechResponse.ID,
	}
	fmt.Printf("Sending message hate status %d\n", hateSpeechResponse.ID)
	go uc.messagingServer.PublishToGroup(
		hateSpeechResponse.GroupID,
		messageDTO.MessageSignal{
			Type:    "hateUpdate",
			Payload: hateMessagecknowledgementPayload,
		},
	)
}
