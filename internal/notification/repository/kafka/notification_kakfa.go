package kafka

import (
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/lightlink/group-service/internal/notification/domain/dto"
	"github.com/linkedin/goavro"
	"github.com/riferrei/srclient"
)

type NotificationKafkaRepository struct {
	producer *kafka.Producer
	codec    *goavro.Codec
	topic    string
}

func NewNotificationKafkaRepository(brokers, topic, schemaRegistryURL string) (*NotificationKafkaRepository, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания продюсера Kafka: %v", err)
	}

	schemaRegistryClient := srclient.CreateSchemaRegistryClient(schemaRegistryURL)
	subject := topic + "-value"
	schema, err := schemaRegistryClient.GetLatestSchema(subject)
	if err != nil {
		log.Println("Схема не найдена, создаём новую...")

		schemaStr := `{
			"type": "record",
			"name": "RawNotification",
			"fields": [
				{"name": "type", "type": "string"},
				{"name": "payload", "type": [
					{"type": "record", "name": "FriendRequestPayload", "fields": [
						{"name": "from_user_id", "type": "string"},
						{"name": "to_user_id", "type": "string"}
					]},
					{"type": "record", "name": "IncomingMessagePayload", "fields": [
						{"name": "from_user_id", "type": "string"},
						{"name": "to_user_id", "type": "string"},
						{"name": "room_id", "type": "string"},
						{"name": "content", "type": "string"}
					]},
					{"type": "record", "name": "IncomingCallPayload", "fields": [
						{"name": "from_user_id", "type": "string"},
						{"name": "to_user_id", "type": "string"},
						{"name": "room_id", "type": "string"}
					]}
				]}
			]
		}`

		schema, err = schemaRegistryClient.CreateSchema(subject, schemaStr, srclient.Avro)
		if err != nil {
			return nil, fmt.Errorf("ошибка создания схемы в Registry: %v", err)
		}
		log.Println("Схема успешно создана!")
	}

	codec, err := goavro.NewCodec(schema.Schema())
	if err != nil {
		return nil, fmt.Errorf("ошибка создания Avro-кодека: %v", err)
	}

	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Ошибка продюсинга: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Сообщение отправлено в %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	return &NotificationKafkaRepository{
		producer: producer,
		codec:    codec,
		topic:    topic,
	}, nil
}

func (repo *NotificationKafkaRepository) Send(notification dto.RawNotification) error {
	var payload map[string]interface{}

	switch notification.Type {
	case "friendRequest":
		payload = map[string]interface{}{
			"FriendRequestPayload": notification.Payload,
		}
	case "incomingMessage":
		payload = map[string]interface{}{
			"IncomingMessagePayload": notification.Payload,
		}
	case "incomingCall":
		payload = map[string]interface{}{
			"IncomingCallPayload": notification.Payload,
		}
	default:
		return fmt.Errorf("неизвестный тип уведомления: %s", notification.Type)
	}

	avroData, err := repo.codec.BinaryFromNative(nil, map[string]interface{}{
		"type":    notification.Type,
		"payload": payload,
	})
	if err != nil {
		return fmt.Errorf("ошибка кодирования Avro: %v", err)
	}

	err = repo.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &repo.topic, Partition: kafka.PartitionAny},
		Value:          avroData,
	}, nil)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения в Kafka: %v", err)
	}

	fmt.Printf("KAFKA: Send value in queue: type-%s, payload-%v\n", notification.Type, payload)

	return nil
}
