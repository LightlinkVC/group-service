package dto

import "github.com/lightlink/group-service/internal/group/domain/entity"

type CreateGroupRequest struct {
	UserID    uint   `json:"user_id"`
	GroupName string `json:"group_name"`
	Type      string `json:"group_type"`
}

func CreateGroupRequestToEntity(createRequest *CreateGroupRequest) *entity.Group {
	return &entity.Group{
		Name:      createRequest.GroupName,
		CreatorID: createRequest.UserID,
		TypeName:  createRequest.Type,
	}
}
