package grpc

import (
	"social-service/internal/producer"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		mock.ExpectClose()
		_ = db.Close()
	}()

	var testProducer *producer.RatingProducer

	s := Init(db, testProducer)

	assert.NotNil(t, s)
	defer s.Stop()
}
