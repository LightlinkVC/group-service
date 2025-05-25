package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/lightlink/group-service/internal/message/domain/dto"
	"github.com/lightlink/group-service/internal/message/usecase"
)

type MessageFilterConsumer struct {
	consumer  *kafka.Consumer
	messageUC usecase.MessageUsecaseI
}

func NewMessageFilterConsumer(messageUC usecase.MessageUsecaseI, brokers, groupID, topic string) (*MessageFilterConsumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return nil, err
	}

	return &MessageFilterConsumer{
		consumer:  consumer,
		messageUC: messageUC,
	}, nil
}

func (c *MessageFilterConsumer) Receive() {
	for {
		timeoutDuration := time.Second * 5

		msg, err := c.consumer.ReadMessage(timeoutDuration)
		if err != nil {
			kafkaErr := err.(kafka.Error)
			if kafkaErr.Code() == kafka.ErrTimedOut {
				continue
			}
			log.Printf("Ошибка при чтении сообщения: %v", err)
			continue
		}

		var response dto.MessageHateSpeechResponse
		if err := json.Unmarshal(msg.Value, &response); err != nil {
			log.Printf("Ошибка декодирования JSON: %v", err)
			continue
		}

		fmt.Printf("KAFKA: Получен результат: %+v\n", response)

		c.messageUC.UpdateHateSpeechLabel(response)
	}
}
