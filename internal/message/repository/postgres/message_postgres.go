package postgres

import (
	"database/sql"
	"fmt"

	"github.com/lightlink/group-service/internal/message/domain/entity"
)

type MessagePostgresRepository struct {
	DB *sql.DB
}

func NewMessagePostgresRepository(db *sql.DB) *MessagePostgresRepository {
	return &MessagePostgresRepository{
		DB: db,
	}
}

func (repo *MessagePostgresRepository) Create(messageEntity *entity.Message) (*entity.Message, error) {
	createdMessageEntity := entity.Message{}

	var messageID uint
	err := repo.DB.QueryRow(`
		INSERT INTO messages (user_id, group_id, content, status_id)
		VALUES ($1, $2, $3, (SELECT id FROM message_statuses WHERE name = 'pending'))
		RETURNING id`,
		messageEntity.UserID, messageEntity.GroupID, messageEntity.Content,
	).Scan(&messageID)
	if err != nil {
		fmt.Println("Create message err")
		return nil, err
	}

	err = repo.DB.QueryRow(`
		SELECT m.id, m.user_id, m.group_id, ms.name, m.content, m.created_at
		FROM messages m
		JOIN message_statuses ms ON m.status_id = ms.id
		WHERE m.id = $1`,
		messageID,
	).Scan(
		&createdMessageEntity.ID,
		&createdMessageEntity.UserID,
		&createdMessageEntity.GroupID,
		&createdMessageEntity.Status,
		&createdMessageEntity.Content,
		&createdMessageEntity.CreatedAt,
	)
	if err != nil {
		fmt.Println("Parse created message error")
		return nil, err
	}

	return &createdMessageEntity, nil
}

func (repo *MessagePostgresRepository) GetByGroupID(groupID uint) ([]entity.Message, error) {
	rows, err := repo.DB.Query(
		`SELECT m.id, m.user_id, m.group_id, ms.name, m.content, m.created_at
		FROM messages m
		JOIN message_statuses ms ON m.status_id = ms.id
		WHERE m.group_id = $1
		ORDER BY m.created_at ASC`,
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
			&currentGroupMessage.Status,
			&currentGroupMessage.Content,
			&currentGroupMessage.CreatedAt,
		)
		if err != nil {
			fmt.Println("Failed to select messages")
			return nil, err
		}

		groupMessages = append(groupMessages, currentGroupMessage)
	}

	return groupMessages, nil
}

func (repo *MessagePostgresRepository) UpdateStatus(messageID uint, statusName string) error {
	_, err := repo.DB.Exec(`
		UPDATE messages 
		SET status_id = (SELECT id FROM message_statuses WHERE name = $1)
		WHERE id = $2
	`, statusName, messageID)

	if err != nil {
		fmt.Printf("Failed to update message status: %v\n", err)
		return err
	}

	return nil
}
