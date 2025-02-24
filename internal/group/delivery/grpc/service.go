package grpc

import (
	"context"

	"github.com/lightlink/group-service/internal/group/domain/dto"
	"github.com/lightlink/group-service/internal/group/domain/entity"
	"github.com/lightlink/group-service/internal/group/usecase"
	proto "github.com/lightlink/group-service/protogen/group"
)

type GroupService struct {
	proto.UnimplementedGroupServiceServer
	groupUC usecase.GroupUsecaseI
}

func NewGroupService(uc usecase.GroupUsecaseI) *GroupService {
	return &GroupService{
		groupUC: uc,
	}
}

func (gs *GroupService) CreatePersonalGroup(ctx context.Context, createRequest *proto.CreatePersonalGroupRequest) (*proto.CreatePersonalGroupResponse, error) {

	groupEntity := dto.CreatePersonalGroupRequestToEntity(createRequest)
	err := gs.groupUC.Create(groupEntity, []entity.GroupMember{
		{
			UserID: uint(createRequest.User1Id),
			Role:   "admin",
		},
		{
			UserID: uint(createRequest.User2Id),
			Role:   "admin",
		},
	})
	if err != nil {
		return nil, err
	}

	return &proto.CreatePersonalGroupResponse{Status: true}, nil
}
