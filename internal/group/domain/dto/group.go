package dto

import (
	"fmt"

	"github.com/lightlink/group-service/internal/group/domain/entity"
	proto "github.com/lightlink/group-service/protogen/group"
)

type GetPersonalGroupIDResponse struct {
	GroupID uint `json:"group_id"`
}

type GetGroupResponse struct {
	GroupID   uint   `json:"group_id"`
	GroupName string `json:"name"`
}

type CreateGroupRequest struct {
	Name    string           `json:"name"`
	Members []GroupMemberDTO `json:"members"`
}

type GroupMemberDTO struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
}

func CreatePersonalGroupRequestToEntity(createRequest *proto.CreatePersonalGroupRequest) *entity.Group {
	return &entity.Group{
		Name:      fmt.Sprintf("personal-%d-%d", createRequest.User1Id, createRequest.User2Id),
		CreatorID: uint(createRequest.User1Id),
		TypeName:  "personal",
	}
}
