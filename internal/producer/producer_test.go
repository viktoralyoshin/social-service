package producer

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestRatingProducer_Publish(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockWriter := new(MockKafkaWriter)
		producer := &RatingProducer{writer: mockWriter}

		gameID := uuid.New()
		ctx := context.Background()

		mockWriter.On("WriteMessages", ctx, mock.MatchedBy(func(msgs []kafka.Message) bool {
			return len(msgs) == 1 && assert.Contains(t, string(msgs[0].Value), gameID.String())
		})).Return(nil).Once()

		err := producer.Publish(ctx, gameID)

		assert.NoError(t, err)
		mockWriter.AssertExpectations(t)
	})

	t.Run("kafka write error", func(t *testing.T) {
		mockWriter := new(MockKafkaWriter)
		producer := &RatingProducer{writer: mockWriter}

		gameID := uuid.New()
		mockWriter.On("WriteMessages", mock.Anything, mock.Anything).
			Return(errors.New("connection reset")).Once()

		err := producer.Publish(context.Background(), gameID)

		assert.Error(t, err)
		assert.Equal(t, "connection reset", err.Error())
	})
}

func TestRatingProducer_Close(t *testing.T) {
	mockWriter := new(MockKafkaWriter)
	producer := &RatingProducer{writer: mockWriter}

	mockWriter.On("Close").Return(nil).Once()
	err := producer.Close()
	assert.NoError(t, err)
}

func TestNewRatingProducer(t *testing.T) {
	p := NewRatingProducer("localhost:9092", "test-topic")
	assert.NotNil(t, p)
	assert.NotNil(t, p.writer)

	_ = p.Close()
}
