package service

import (
	"context"
	"social-service/internal/storage"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	socialpb "github.com/viktoralyoshin/playhub-proto/gen/go/social"
)

func setupServiceTest(t *testing.T) (*ReviewService, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := storage.NewReviewRepo(db)
	svc := NewReviewService(repo)

	return svc, mock, func() {
		_ = db.Close()
	}
}

func TestReviewService_CreateReview(t *testing.T) {
	svc, mock, cleanup := setupServiceTest(t)
	defer cleanup()

	req := &socialpb.CreateReviewRequest{
		UserId: uuid.New().String(),
		GameId: uuid.New().String(),
		Rating: 5,
		Text:   "Great",
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "game_id", "rating", "text", "created_at", "updated_at"}).
			AddRow(uuid.New().String(), req.UserId, req.GameId, req.Rating, req.Text, time.Now(), time.Now())

		mock.ExpectQuery(`INSERT INTO social.reviews`).WillReturnRows(rows)

		res, err := svc.CreateReview(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}

func TestReviewService_GetReviewsByUser(t *testing.T) {
	svc, mock, cleanup := setupServiceTest(t)
	defer cleanup()

	userID := uuid.New().String()

	t.Run("normalization of negative values", func(t *testing.T) {
		req := &socialpb.GetUserReviewsRequest{
			UserId: userID,
			Limit:  -10,
			Offset: -5,
		}

		mock.ExpectQuery(`LIMIT \$2 OFFSET \$3`).
			WithArgs(userID, 0, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := svc.GetReviewsByUser(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, int32(0), req.Limit)
		assert.Equal(t, int32(0), req.Offset)
	})

	t.Run("positive values remains", func(t *testing.T) {
		req := &socialpb.GetUserReviewsRequest{UserId: userID, Limit: 10, Offset: 20}
		mock.ExpectQuery(`LIMIT \$2 OFFSET \$3`).
			WithArgs(userID, 10, 20).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := svc.GetReviewsByUser(context.Background(), req)
		assert.NoError(t, err)
	})
}

func TestReviewService_GetFeed(t *testing.T) {
	svc, mock, cleanup := setupServiceTest(t)
	defer cleanup()

	t.Run("normalization of negative limit", func(t *testing.T) {
		req := &socialpb.GetFeedRequest{Limit: -1}

		mock.ExpectQuery(`LIMIT \$1`).
			WithArgs(0).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := svc.GetFeed(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, int32(0), req.Limit)
	})
}

func TestReviewService_GetReviewsByGame(t *testing.T) {
	svc, mock, cleanup := setupServiceTest(t)
	defer cleanup()

	gameID := uuid.New().String()

	t.Run("normalization of negative values", func(t *testing.T) {
		req := &socialpb.GetGameReviewsRequest{
			GameId: gameID,
			Limit:  -100,
			Offset: -1,
		}

		mock.ExpectQuery(`WHERE game_id = \$1`).
			WithArgs(gameID, nil, 0).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := svc.GetReviewsByGame(context.Background(), req)
		assert.NoError(t, err)
	})

	t.Run("success with positive values", func(t *testing.T) {
		req := &socialpb.GetGameReviewsRequest{GameId: gameID, Limit: 5, Offset: 10}

		mock.ExpectQuery(`WHERE game_id = \$1`).
			WithArgs(gameID, 5, 10).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		_, err := svc.GetReviewsByGame(context.Background(), req)
		assert.NoError(t, err)
	})
}
