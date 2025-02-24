package usecase

import (
	"github.com/lightlink/group-service/internal/message/domain/dto"
	"github.com/lightlink/group-service/internal/message/domain/entity"
	"github.com/lightlink/group-service/internal/message/repository"
)

type MessageUsecaseI interface {
	Create(createRequest *dto.CreateMessageRequest) error
	GetByGroupID(groupID uint) ([]entity.Message, error)
}

type MessageUsecase struct {
	messageRepo repository.MessageRepositoryI
}

func NewMessageUsecase(messageRepo repository.MessageRepositoryI) *MessageUsecase {
	return &MessageUsecase{
		messageRepo: messageRepo,
	}
}

func (uc *MessageUsecase) Create(createRequest *dto.CreateMessageRequest) error {
	messageEntity := entity.Message{
		UserID:  createRequest.UserID,
		GroupID: createRequest.GroupID,
		Content: createRequest.Content,
	}

	_, err := uc.messageRepo.Create(&messageEntity)
	if err != nil {
		return err
	}

	return nil
}

func (uc *MessageUsecase) GetByGroupID(groupID uint) ([]entity.Message, error) {
	messages, err := uc.messageRepo.GetByGroupID(groupID)
	if err != nil {
		return nil, err
	}

	return messages, nil
}
