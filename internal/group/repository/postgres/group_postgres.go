package postgres

import (
	"database/sql"

	"github.com/lightlink/group-service/internal/group/domain/entity"
	"github.com/lightlink/group-service/internal/group/domain/model"
)

type GroupPostgresRepository struct {
	DB *sql.DB
}

func NewGroupPostgresRepository(db *sql.DB) *GroupPostgresRepository {
	return &GroupPostgresRepository{
		DB: db,
	}
}

func (repo *GroupPostgresRepository) Create(groupEntity *entity.Group) (*model.Group, error) {
	createdGroupModel := &model.Group{}

	var groupTypeID uint

	err := repo.DB.QueryRow(
		"SELECT id FROM group_types WHERE name = $1",
		groupEntity.TypeName,
	).Scan(&groupTypeID)
	if err != nil {
		return nil, err
	}

	err = repo.DB.QueryRow(
		`INSERT INTO groups (name, creator_id, type_id) 
		VALUES ($1, $2, $3) 
		RETURNING id, name, creator_id, type_id`,
		groupEntity.Name, groupEntity.CreatorID, groupTypeID,
	).Scan(
		&createdGroupModel.ID,
		&createdGroupModel.Name,
		&createdGroupModel.CreatorID,
		&createdGroupModel.TypeID,
	)
	if err != nil {
		return nil, err
	}

	return createdGroupModel, nil
}
