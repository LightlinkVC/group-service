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
	tx, err := repo.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var messageID uint
	err = tx.QueryRow(`
        INSERT INTO messages (user_id, group_id, content, status_id)
        VALUES ($1, $2, $3, (SELECT id FROM message_statuses WHERE name = 'pending'))
        RETURNING id`,
		messageEntity.UserID, messageEntity.GroupID, messageEntity.Content,
	).Scan(&messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert message: %w", err)
	}

	for _, file := range messageEntity.Files {
		_, err = tx.Exec(`
            INSERT INTO files (message_id, object_name, original_name, content_type, size, url)
            VALUES ($1, $2, $3, $4, $5, $6)`,
			messageID,
			file.ObjectName,
			file.OriginalName,
			file.ContentType,
			file.Size,
			file.URL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to insert file: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return repo.getMessageWithFiles(messageID)
}

func (repo *MessagePostgresRepository) getMessageWithFiles(messageID uint) (*entity.Message, error) {
	message := &entity.Message{}

	err := repo.DB.QueryRow(`
        SELECT m.id, m.user_id, m.group_id, ms.name, m.content, m.created_at
        FROM messages m
        JOIN message_statuses ms ON m.status_id = ms.id
        WHERE m.id = $1`,
		messageID,
	).Scan(
		&message.ID,
		&message.UserID,
		&message.GroupID,
		&message.Status,
		&message.Content,
		&message.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	// Получаем файлы
	files, err := repo.getFilesByMessageID(messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}
	message.Files = files

	return message, nil
}

func (repo *MessagePostgresRepository) getFilesByMessageID(messageID uint) ([]entity.File, error) {
	rows, err := repo.DB.Query(`
        SELECT id, object_name, original_name, content_type, size, url
        FROM files
        WHERE message_id = $1`,
		messageID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var files []entity.File
	for rows.Next() {
		var f entity.File
		if err := rows.Scan(
			&f.ID,
			&f.ObjectName,
			&f.OriginalName,
			&f.ContentType,
			&f.Size,
			&f.URL,
		); err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}
		files = append(files, f)
	}

	return files, nil
}

func (repo *MessagePostgresRepository) getMessagesByGroupID(groupID uint) ([]entity.Message, error) {
	rows, err := repo.DB.Query(`
        SELECT m.id, m.user_id, m.group_id, ms.name, m.content, m.created_at
        FROM messages m
        JOIN message_statuses ms ON m.status_id = ms.id
        WHERE m.group_id = $1
        ORDER BY m.created_at ASC`,
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []entity.Message
	for rows.Next() {
		var msg entity.Message
		if err := rows.Scan(
			&msg.ID,
			&msg.UserID,
			&msg.GroupID,
			&msg.Status,
			&msg.Content,
			&msg.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (repo *MessagePostgresRepository) GetByGroupID(groupID uint) ([]entity.Message, error) {
	// Получаем основные данные сообщений
	messages, err := repo.getMessagesByGroupID(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Для каждого сообщения получаем файлы
	for i := range messages {
		files, err := repo.getFilesByMessageID(messages[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get files for message %d: %w", messages[i].ID, err)
		}
		messages[i].Files = files
	}

	return messages, nil
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
