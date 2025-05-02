package repository

import "github.com/lightlink/group-service/internal/message/domain/dto"

type MessageHateSpeechRepositoryI interface {
	Send(hateSpeechRequest dto.MessageHateSpeechRequest) error
}
