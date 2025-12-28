package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	socialpb "github.com/viktoralyoshin/playhub-proto/gen/go/social"
	"github.com/viktoralyoshin/utils/pkg/errs"
)

func setupReviewRepoTest(t *testing.T) (*ReviewRepo, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := NewReviewRepo(db)

	return repo, mock, func() {
		_ = db.Close()
	}
}

func TestReviewRepo_CreateReview(t *testing.T) {
	repo, mock, cleanup := setupReviewRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	req := &socialpb.CreateReviewRequest{
		UserId: uuid.New().String(),
		GameId: uuid.New().String(),
		Rating: 100,
		Text:   "Great game!",
	}

	columns := []string{"id", "user_id", "game_id", "rating", "text", "created_at", "updated_at"}

	t.Run("success", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows(columns).
			AddRow(uuid.New().String(), req.UserId, req.GameId, req.Rating, req.Text, now, now)

		mock.ExpectQuery(`INSERT INTO social.reviews`).
			WithArgs(req.UserId, req.GameId, req.Rating, req.Text).
			WillReturnRows(rows)

		res, err := repo.CreateReview(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int(req.Rating), res.Rating)
	})

	t.Run("duplicate review error", func(t *testing.T) {
		pqErr := &pq.Error{Code: "23505"}
		mock.ExpectQuery(`INSERT INTO social.reviews`).WillReturnError(pqErr)
		res, err := repo.CreateReview(ctx, req)
		assert.ErrorIs(t, err, errs.ErrReviewExists)
		assert.Nil(t, res)
	})

	t.Run("generic db error", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO social.reviews`).WillReturnError(errors.New("db fail"))
		res, err := repo.CreateReview(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestReviewRepo_GetReviewsByUser(t *testing.T) {
	repo, mock, cleanup := setupReviewRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	userID := uuid.New().String()
	req := &socialpb.GetUserReviewsRequest{UserId: userID, Limit: 10, Offset: 0}
	columns := []string{"id", "user_id", "game_id", "rating", "text", "created_at", "updated_at"}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows(columns).
			AddRow(uuid.New().String(), userID, uuid.New().String(), 80, "Nice", time.Now(), time.Now())
		mock.ExpectQuery(`SELECT (.+) FROM social.reviews WHERE user_id = \$1`).WillReturnRows(rows)
		res, err := repo.GetReviewsByUser(ctx, req)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT`).WillReturnError(errors.New("fail"))
		_, err := repo.GetReviewsByUser(ctx, req)
		assert.Error(t, err)
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).AddRow("not-uuid")
		mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
		_, err := repo.GetReviewsByUser(ctx, req)
		assert.Error(t, err)
	})

	t.Run("rows error", func(t *testing.T) {
		rows := sqlmock.NewRows(columns).AddRow(uuid.New().String(), userID, uuid.New().String(), 80, "Nice", time.Now(), time.Now()).
			RowError(0, errors.New("iteration error"))
		mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
		_, err := repo.GetReviewsByUser(ctx, req)
		assert.Error(t, err)
	})
}

func TestReviewRepo_GetFeed(t *testing.T) {
	repo, mock, cleanup := setupReviewRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	req := &socialpb.GetFeedRequest{Limit: 5}
	columns := []string{"id", "user_id", "game_id", "rating", "text", "created_at", "updated_at"}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows(columns).
			AddRow(uuid.New().String(), uuid.New().String(), uuid.New().String(), 50, "T1", time.Now(), time.Now())
		mock.ExpectQuery(`SELECT (.+) FROM social.reviews ORDER BY created_at DESC LIMIT \$1`).WillReturnRows(rows)
		res, err := repo.GetFeed(ctx, req)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT`).WillReturnError(errors.New("db error"))
		_, err := repo.GetFeed(ctx, req)
		assert.Error(t, err)
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).AddRow("not-uuid")
		mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
		_, err := repo.GetFeed(ctx, req)
		assert.Error(t, err)
	})

	t.Run("rows error", func(t *testing.T) {
		rows := sqlmock.NewRows(columns).AddRow(uuid.New().String(), uuid.New().String(), uuid.New().String(), 5, "T", time.Now(), time.Now()).
			RowError(0, errors.New("stream error"))
		mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
		_, err := repo.GetFeed(ctx, req)
		assert.Error(t, err)
	})
}

func TestReviewRepo_GetReviewsByGame(t *testing.T) {
	repo, mock, cleanup := setupReviewRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	gameID := uuid.New().String()
	columns := []string{"id", "user_id", "game_id", "rating", "text", "created_at", "updated_at"}

	t.Run("success with limit", func(t *testing.T) {
		req := &socialpb.GetGameReviewsRequest{GameId: gameID, Limit: 10, Offset: 0}
		rows := sqlmock.NewRows(columns).
			AddRow(uuid.New().String(), uuid.New().String(), gameID, 50, "T1", time.Now(), time.Now())
		mock.ExpectQuery(`WHERE game_id = \$1`).WillReturnRows(rows)
		res, err := repo.GetReviewsByGame(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("query error", func(t *testing.T) {
		req := &socialpb.GetGameReviewsRequest{GameId: gameID, Limit: 10}
		mock.ExpectQuery(`SELECT`).WillReturnError(errors.New("db fail"))
		_, err := repo.GetReviewsByGame(ctx, req)
		assert.Error(t, err)
	})

	t.Run("scan error", func(t *testing.T) {
		req := &socialpb.GetGameReviewsRequest{GameId: gameID, Limit: 10}
		rows := sqlmock.NewRows([]string{"id"}).AddRow("not-uuid")
		mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
		_, err := repo.GetReviewsByGame(ctx, req)
		assert.Error(t, err)
	})

	t.Run("rows error during iteration", func(t *testing.T) {
		req := &socialpb.GetGameReviewsRequest{GameId: gameID, Limit: 10}
		rows := sqlmock.NewRows(columns).AddRow(uuid.New().String(), uuid.New().String(), gameID, 5, "T", time.Now(), time.Now()).
			RowError(0, errors.New("broken pipe"))
		mock.ExpectQuery(`SELECT`).WillReturnRows(rows)
		_, err := repo.GetReviewsByGame(ctx, req)
		assert.Error(t, err)
	})

	t.Run("success with zero limit", func(t *testing.T) {
		req := &socialpb.GetGameReviewsRequest{GameId: gameID, Limit: 0, Offset: 0}
		mock.ExpectQuery(`WHERE game_id = \$1`).WithArgs(gameID, nil, 0).WillReturnRows(sqlmock.NewRows(columns))
		_, err := repo.GetReviewsByGame(ctx, req)
		assert.NoError(t, err)
	})
}
