package handlers

import (
	"context"
	"errors"
	"social-service/internal/microservice"
	"social-service/internal/service"
	"social-service/internal/storage"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	authpb "github.com/viktoralyoshin/playhub-proto/gen/go/auth"
	gamepb "github.com/viktoralyoshin/playhub-proto/gen/go/games"
	socialpb "github.com/viktoralyoshin/playhub-proto/gen/go/social"
	"github.com/viktoralyoshin/utils/pkg/errs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) Publish(ctx context.Context, gameID uuid.UUID) error {
	args := m.Called(ctx, gameID)
	return args.Error(0)
}

type MockAuthClient struct {
	mock.Mock
	authpb.AuthServiceClient
}

func (m *MockAuthClient) GetUser(ctx context.Context, in *authpb.GetUserRequest, opts ...grpc.CallOption) (*authpb.GetUserResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*authpb.GetUserResponse), args.Error(1)
}

type MockGamesClient struct {
	mock.Mock
	gamepb.GameServiceClient
}

func (m *MockGamesClient) GetGame(ctx context.Context, in *gamepb.GetGameRequest, opts ...grpc.CallOption) (*gamepb.GetGameResponse, error) {
	args := m.Called(ctx, in)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gamepb.GetGameResponse), args.Error(1)
}

func setupHandlerTest(t *testing.T) (*ReviewHandler, *MockProducer, sqlmock.Sqlmock, func()) {
	db, dbMock, err := sqlmock.New()
	require.NoError(t, err)

	repo := storage.NewReviewRepo(db)
	svc := service.NewReviewService(repo)

	mockProd := new(MockProducer)
	h := NewReviewHandler(svc, mockProd)

	return h, mockProd, dbMock, func() {
		dbMock.ExpectClose()
		_ = db.Close()
	}
}

func TestReviewHandler_CreateReview(t *testing.T) {
	h, mockProd, dbMock, cleanup := setupHandlerTest(t)
	defer cleanup()

	userID := uuid.New()
	gameID := uuid.New()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-user-id", userID.String()))
	req := &socialpb.CreateReviewRequest{GameId: gameID.String(), Rating: 5, Text: "Great!"}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "game_id", "rating", "text", "created_at", "updated_at"}).
			AddRow(uuid.New().String(), userID.String(), gameID.String(), 5, "Great!", time.Now(), time.Now())

		dbMock.ExpectQuery(`INSERT INTO`).WillReturnRows(rows)
		mockProd.On("Publish", mock.Anything, gameID).Return(nil).Once()

		resp, err := h.CreateReview(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("permission denied - no metadata", func(t *testing.T) {
		_, err := h.CreateReview(context.Background(), req)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	t.Run("already exists", func(t *testing.T) {
		dbMock.ExpectQuery(`INSERT INTO`).WillReturnError(errs.ErrReviewExists)
		_, err := h.CreateReview(ctx, req)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})

	t.Run("internal service error", func(t *testing.T) {
		dbMock.ExpectQuery(`INSERT INTO`).WillReturnError(errors.New("db fail"))
		_, err := h.CreateReview(ctx, req)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("producer failure - still success response", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "game_id", "rating", "text", "created_at", "updated_at"}).
			AddRow(uuid.New().String(), userID.String(), gameID.String(), 5, "Great!", time.Now(), time.Now())

		dbMock.ExpectQuery(`INSERT INTO`).WillReturnRows(rows)
		mockProd.On("Publish", mock.Anything, gameID).Return(errors.New("kafka error")).Once()

		resp, err := h.CreateReview(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}

func TestReviewHandler_GetFeed(t *testing.T) {
	h, _, dbMock, cleanup := setupHandlerTest(t)
	defer cleanup()

	t.Run("success with rows mapping", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "game_id", "rating", "text", "created_at", "updated_at"}).
			AddRow(uuid.New().String(), uuid.New().String(), uuid.New().String(), 5, "T1", time.Now(), time.Now()).
			AddRow(uuid.New().String(), uuid.New().String(), uuid.New().String(), 4, "T2", time.Now(), time.Now())

		dbMock.ExpectQuery(`SELECT`).WillReturnRows(rows)
		resp, err := h.GetFeed(context.Background(), &socialpb.GetFeedRequest{Limit: 2})
		assert.NoError(t, err)
		assert.Len(t, resp.Reviews, 2)
	})

	t.Run("service error", func(t *testing.T) {
		dbMock.ExpectQuery(`SELECT`).WillReturnError(errors.New("fail"))
		_, err := h.GetFeed(context.Background(), &socialpb.GetFeedRequest{Limit: 1})
		assert.Equal(t, codes.Internal, status.Code(err))
	})
}

func TestReviewHandler_GetUserReviews(t *testing.T) {
	h, _, dbMock, cleanup := setupHandlerTest(t)
	defer cleanup()

	authMock := new(MockAuthClient)
	oldAuth := microservice.AuthClient
	microservice.AuthClient = authMock
	defer func() { microservice.AuthClient = oldAuth }()

	targetUID := uuid.New().String()

	t.Run("user not found in auth-service", func(t *testing.T) {
		authMock.On("GetUser", mock.Anything, mock.Anything).Return(nil, errors.New("grpc error")).Once()
		_, err := h.GetUserReviews(context.Background(), &socialpb.GetUserReviewsRequest{UserId: targetUID})
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("internal service error", func(t *testing.T) {
		authMock.On("GetUser", mock.Anything, mock.Anything).Return(&authpb.GetUserResponse{}, nil).Once()
		dbMock.ExpectQuery(`SELECT`).WillReturnError(errors.New("db error"))

		_, err := h.GetUserReviews(context.Background(), &socialpb.GetUserReviewsRequest{UserId: targetUID})
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("success", func(t *testing.T) {
		authMock.On("GetUser", mock.Anything, mock.Anything).Return(&authpb.GetUserResponse{}, nil).Once()
		rows := sqlmock.NewRows([]string{"id", "user_id", "game_id", "rating", "text", "created_at", "updated_at"}).
			AddRow(uuid.New().String(), targetUID, uuid.New().String(), 5, "T", time.Now(), time.Now())
		dbMock.ExpectQuery(`SELECT`).WillReturnRows(rows)

		resp, err := h.GetUserReviews(context.Background(), &socialpb.GetUserReviewsRequest{UserId: targetUID})
		assert.NoError(t, err)
		assert.Len(t, resp.Reviews, 1)
	})
}

func TestReviewHandler_GetGameReviews(t *testing.T) {
	h, _, dbMock, cleanup := setupHandlerTest(t)
	defer cleanup()

	gamesMock := new(MockGamesClient)
	oldGames := microservice.GamesClient
	microservice.GamesClient = gamesMock
	defer func() { microservice.GamesClient = oldGames }()

	gameID := uuid.New().String()

	t.Run("game not found in games-service", func(t *testing.T) {
		gamesMock.On("GetGame", mock.Anything, mock.Anything).Return(nil, errors.New("grpc error")).Once()
		_, err := h.GetGameReviews(context.Background(), &socialpb.GetGameReviewsRequest{GameId: gameID})
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("internal service error", func(t *testing.T) {
		gamesMock.On("GetGame", mock.Anything, mock.Anything).Return(&gamepb.GetGameResponse{}, nil).Once()
		dbMock.ExpectQuery(`SELECT`).WillReturnError(errors.New("db error"))

		_, err := h.GetGameReviews(context.Background(), &socialpb.GetGameReviewsRequest{GameId: gameID})
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("success", func(t *testing.T) {
		gamesMock.On("GetGame", mock.Anything, mock.Anything).Return(&gamepb.GetGameResponse{}, nil).Once()
		rows := sqlmock.NewRows([]string{"id", "user_id", "game_id", "rating", "text", "created_at", "updated_at"}).
			AddRow(uuid.New().String(), uuid.New().String(), gameID, 5, "T", time.Now(), time.Now())
		dbMock.ExpectQuery(`SELECT`).WillReturnRows(rows)

		resp, err := h.GetGameReviews(context.Background(), &socialpb.GetGameReviewsRequest{GameId: gameID})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})
}
