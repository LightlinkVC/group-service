package postgres

import (
	"database/sql"
	"fmt"

	"github.com/lightlink/group-service/internal/message/domain/entity"
	"github.com/lightlink/group-service/internal/message/domain/model"
)

type MessagePostgresRepository struct {
	DB *sql.DB
}

func NewMessagePostgresRepository(db *sql.DB) *MessagePostgresRepository {
	return &MessagePostgresRepository{
		DB: db,
	}
}

func (repo *MessagePostgresRepository) Create(messageEntity *entity.Message) (*model.Message, error) {
	createdMessageModel := model.Message{}

	err := repo.DB.QueryRow(
		`INSERT INTO messages (user_id, group_id, content) 
		VALUES ($1, $2, $3) 
		RETURNING id, user_id, group_id, content`,
		messageEntity.UserID, messageEntity.GroupID, messageEntity.Content,
	).Scan(
		&createdMessageModel.ID,
		&createdMessageModel.UserID,
		&createdMessageModel.GroupID,
		&createdMessageModel.Content,
	)
	if err != nil {
		fmt.Println("Create message err")
		return nil, err
	}

	return &createdMessageModel, nil
}

func (repo *MessagePostgresRepository) GetByGroupID(groupID uint) ([]entity.Message, error) {
	rows, err := repo.DB.Query(
		`SELECT id, user_id, group_id, content
		FROM messages m
		WHERE m.group_id = $1`,
		groupID,
	)
	if err != nil {
		fmt.Println("Failed to select messages by group id")
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Println("Failed to close rows selecting messages")
		}
	}()

	groupMessages := []entity.Message{}
	for rows.Next() {
		currentGroupMessage := entity.Message{}
		err := rows.Scan(
			&currentGroupMessage.ID,
			&currentGroupMessage.UserID,
			&currentGroupMessage.GroupID,
			&currentGroupMessage.Content,
		)
		if err != nil {
			fmt.Println("Failed to select messages")
			return nil, err
		}

		groupMessages = append(groupMessages, currentGroupMessage)
	}

	return groupMessages, nil
}
