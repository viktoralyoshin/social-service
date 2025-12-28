package utils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/viktoralyoshin/utils/pkg/errs"
	"google.golang.org/grpc/metadata"
)

func TestGetUserID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		md := metadata.Pairs("x-user-id", "test-uuid-123")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		userID, err := GetUserID(ctx)

		assert.NoError(t, err)
		assert.Equal(t, "test-uuid-123", userID)
	})

	t.Run("no metadata in context", func(t *testing.T) {
		ctx := context.Background()

		userID, err := GetUserID(ctx)

		assert.Empty(t, userID)
		assert.ErrorIs(t, err, errs.ErrInvalidMetadata)
	})

	t.Run("metadata exists but key is missing", func(t *testing.T) {
		md := metadata.Pairs("authorization", "bearer-token")
		ctx := metadata.NewIncomingContext(context.Background(), md)

		userID, err := GetUserID(ctx)

		assert.Empty(t, userID)
		assert.ErrorIs(t, err, errs.ErrMetadataNotFound)
	})

	t.Run("metadata exists but key is empty", func(t *testing.T) {
		md := metadata.MD{"x-user-id": []string{}}
		ctx := metadata.NewIncomingContext(context.Background(), md)

		userID, err := GetUserID(ctx)

		assert.Empty(t, userID)
		assert.ErrorIs(t, err, errs.ErrMetadataNotFound)
	})
}
