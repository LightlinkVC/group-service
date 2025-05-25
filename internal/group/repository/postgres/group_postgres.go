package postgres

import (
	"database/sql"
	"fmt"

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

func (repo *GroupPostgresRepository) GetPersonalGroupID(user1ID uint, user2ID uint) (uint, error) {
	var groupID uint

	err := repo.DB.QueryRow(
		`SELECT g.id
		FROM groups g
		JOIN group_types gt ON g.type_id = gt.id
		JOIN group_members gm1 ON g.id = gm1.group_id
		JOIN group_members gm2 ON g.id = gm2.group_id
		WHERE gt.name = 'personal' 
			AND gm1.user_id = $1 
			AND gm2.user_id = $2
		LIMIT 1;`,
		user1ID, user2ID).Scan(&groupID)
	if err != nil {
		return 0, err
	}

	return groupID, nil
}

func (repo *GroupPostgresRepository) GetMemberIDsByGroupID(groupID uint) ([]uint, error) {
	var memberIDs []uint

	rows, err := repo.DB.Query(
		"SELECT user_id FROM group_members WHERE group_id = $1",
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query group members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userID uint
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		memberIDs = append(memberIDs, userID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return memberIDs, nil
}

func (repo *GroupPostgresRepository) Create(groupEntity *entity.Group, groupMembers []entity.GroupMember) (*model.Group, error) {
	createdGroupModel := &model.Group{}
	var groupTypeID uint

	err := repo.DB.QueryRow(
		"SELECT id FROM group_types WHERE name = $1",
		groupEntity.TypeName,
	).Scan(&groupTypeID)
	if err != nil {
		return nil, err
	}

	tx, err := repo.DB.Begin()
	if err != nil {
		return nil, err
	}

	err = tx.QueryRow(
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
		fmt.Println("Create group err")
		rbErr := tx.Rollback()
		if rbErr != nil {
			fmt.Println("Rb err")
		}
		return nil, err
	}

	for _, groupMember := range groupMembers {
		var roleID uint

		err := repo.DB.QueryRow(
			"SELECT id FROM roles WHERE name = $1",
			groupMember.Role,
		).Scan(&roleID)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(
			`INSERT INTO group_members (user_id, group_id, role_id) 
			VALUES ($1, $2, $3)`,
			groupMember.UserID, createdGroupModel.ID, roleID,
		)
		if err != nil {
			fmt.Println("Create group member err")
			rbErr := tx.Rollback()
			if rbErr != nil {
				fmt.Println("Rb err")
			}
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println("commit err")
		return nil, err
	}

	return createdGroupModel, nil
}
