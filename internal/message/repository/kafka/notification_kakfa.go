package kafka

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/lightlink/group-service/internal/message/domain/dto"
)

type MessageHateSpeechRepository struct {
	producer *kafka.Producer
	topic    string
}

func NewMessageHateSpeechRepository(brokers, topic string) (*MessageHateSpeechRepository, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания продюсера Kafka: %v", err)
	}

	repo := &MessageHateSpeechRepository{
		producer: producer,
		topic:    topic,
	}

	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Kafka error: %v\n", ev.TopicPartition.Error)
				} else {
					log.Printf("Kafka message delivered to %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	return repo, nil
}

func (repo *MessageHateSpeechRepository) Send(hateSpeechRequest dto.MessageHateSpeechRequest) error {
	hateSpeechRequestJSON, err := json.Marshal(hateSpeechRequest)
	if err != nil {
		return fmt.Errorf("ошибка сериализации hateSpeechRequest: %w", err)
	}

	err = repo.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &repo.topic, Partition: kafka.PartitionAny},
		Value:          hateSpeechRequestJSON,
	}, nil)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения в Kafka: %v", err)
	}

	fmt.Printf("KAFKA: Send value in queue: payload-%v\n", hateSpeechRequest)

	return nil
}
