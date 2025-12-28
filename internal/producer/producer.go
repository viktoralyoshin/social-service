package producer

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type KafkaWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type RatingPublisher interface {
	Publish(ctx context.Context, gameID uuid.UUID) error
}

type ReviewEvent struct {
	GameID string `json:"game_id"`
}

type RatingProducer struct {
	writer KafkaWriter
}

func NewRatingProducer(broker string, topic string) *RatingProducer {
	return &RatingProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
			Async:    false,
		},
	}
}

func (p *RatingProducer) Publish(ctx context.Context, gameId uuid.UUID) error {
	event := &ReviewEvent{
		GameID: gameId.String(),
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Value: body,
	}); err != nil {
		return err
	}

	return nil
}

func (p *RatingProducer) Close() error {
	return p.writer.Close()
}
