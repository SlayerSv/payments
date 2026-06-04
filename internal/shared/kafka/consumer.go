package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers []string, topic string, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID, // Нужен для масштабирования
			MinBytes: 10e3,    // 10KB
			MaxBytes: 10e6,    // 10MB
		}),
	}
}

func (c *Consumer) StartConsume(ctx context.Context, handler func(ctx context.Context, msg []byte) error) {
	go func() {
		for {
			m, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return // Выход при закрытии контекста
				}
				log.Printf("Ошибка чтения сообщения: %v", err)
				continue
			}

			// Вызываем твой обработчик
			if err := handler(ctx, m.Value); err != nil {
				log.Printf("Ошибка обработки сообщения: %v", err)
			}
		}
	}()
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
