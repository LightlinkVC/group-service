package dto

import (
	"fmt"

	"github.com/lightlink/group-service/internal/group/domain/entity"
	proto "github.com/lightlink/group-service/protogen/group"
)

type GetPersonalGroupIDResponse struct {
	GroupID uint `json:"group_id"`
}

func CreatePersonalGroupRequestToEntity(createRequest *proto.CreatePersonalGroupRequest) *entity.Group {
	return &entity.Group{
		Name:      fmt.Sprintf("personal-%d-%d", createRequest.User1Id, createRequest.User2Id),
		CreatorID: uint(createRequest.User1Id),
		TypeName:  "personal",
	}
}
