package handlers

import (
	"context"
	"errors"
	"social-service/internal/microservice"
	"social-service/internal/producer"
	"social-service/internal/service"
	"social-service/internal/utils"

	"github.com/rs/zerolog/log"
	authpb "github.com/viktoralyoshin/playhub-proto/gen/go/auth"
	gamepb "github.com/viktoralyoshin/playhub-proto/gen/go/games"
	socialpb "github.com/viktoralyoshin/playhub-proto/gen/go/social"
	"github.com/viktoralyoshin/utils/pkg/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ReviewHandler struct {
	socialpb.UnimplementedSocialServiceServer
	service  *service.ReviewService
	producer producer.RatingPublisher
}

func NewReviewHandler(service *service.ReviewService, producer producer.RatingPublisher) *ReviewHandler {
	return &ReviewHandler{
		service:  service,
		producer: producer,
	}
}

func (h *ReviewHandler) CreateReview(ctx context.Context, req *socialpb.CreateReviewRequest) (*socialpb.CreateReviewResponse, error) {
	userId, err := utils.GetUserID(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("ReviewHandler.CreateReview: failed to extract user_id from context")
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	req.UserId = userId

	log.Info().
		Str("user_id", userId).
		Str("game_id", req.GameId).
		Int32("rating", req.Rating).
		Msg("ReviewHandler.CreateReview: attempt")

	review, err := h.service.CreateReview(ctx, req)
	if err != nil {
		if errors.Is(err, errs.ErrReviewExists) {
			log.Warn().
				Str("user_id", userId).
				Str("game_id", req.GameId).
				Msg("ReviewHandler.CreateReview: review already exists")
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		log.Error().
			Err(err).
			Str("user_id", userId).
			Str("game_id", req.GameId).
			Msg("ReviewHandler.CreateReview: service error")
		return nil, status.Error(codes.Internal, "internal error during review creation")
	}

	if err := h.producer.Publish(context.Background(), review.GameID); err != nil {
		log.Error().
			Err(err).
			Str("game_id", review.GameID.String()).
			Msg("ReviewHandler.CreateReview: failed to publish rating update to broker")
	} else {
		log.Debug().Str("game_id", review.GameID.String()).Msg("ReviewHandler.CreateReview: rating update published")
	}

	log.Info().
		Str("review_id", review.Id.String()).
		Str("user_id", userId).
		Msg("ReviewHandler.CreateReview: success")

	return &socialpb.CreateReviewResponse{Review: &socialpb.Review{
		Id:        review.Id.String(),
		UserId:    review.UserID.String(),
		GameId:    review.GameID.String(),
		Rating:    int32(review.Rating),
		Text:      review.Text,
		CreatedAt: timestamppb.New(review.CreatedAt),
		UpdatedAt: timestamppb.New(review.CreatedAt),
	}}, nil
}

func (h *ReviewHandler) GetFeed(ctx context.Context, req *socialpb.GetFeedRequest) (*socialpb.GetFeedResponse, error) {
	log.Info().Int32("limit", req.Limit).Msg("ReviewHandler.GetFeed: fetching reviews")

	reviews, err := h.service.GetFeed(ctx, req)
	if err != nil {
		log.Error().Err(err).Msg("ReviewHandler.GetFeed: service error")
		return nil, status.Error(codes.Internal, "failed to get reviews")
	}

	revpb := make([]*socialpb.Review, 0, len(reviews))
	for _, rev := range reviews {
		revpb = append(revpb, &socialpb.Review{
			Id:        rev.Id.String(),
			UserId:    rev.UserID.String(),
			GameId:    rev.GameID.String(),
			Rating:    int32(rev.Rating),
			Text:      rev.Text,
			CreatedAt: timestamppb.New(rev.CreatedAt),
			UpdatedAt: timestamppb.New(rev.UpdatedAt),
		})
	}

	return &socialpb.GetFeedResponse{Reviews: revpb}, nil
}

func (h *ReviewHandler) GetUserReviews(ctx context.Context, req *socialpb.GetUserReviewsRequest) (*socialpb.GetUserReviewsResponse, error) {
	log.Info().Str("target_user_id", req.UserId).Msg("ReviewHandler.GetUserReviews: fetching reviews")

	_, err := microservice.AuthClient.GetUser(ctx, &authpb.GetUserRequest{
		UserId: req.UserId,
	})
	if err != nil {
		log.Warn().
			Err(err).
			Str("target_user_id", req.UserId).
			Msg("ReviewHandler.GetUserReviews: target user check failed (auth-service)")
		return nil, status.Error(codes.NotFound, "user not found")
	}

	reviews, err := h.service.GetReviewsByUser(ctx, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("target_user_id", req.UserId).
			Msg("ReviewHandler.GetUserReviews: service error")
		return nil, status.Error(codes.Internal, "failed to get reviews")
	}

	revpb := make([]*socialpb.Review, 0, len(reviews))
	for _, rev := range reviews {
		revpb = append(revpb, &socialpb.Review{
			Id:        rev.Id.String(),
			UserId:    rev.UserID.String(),
			GameId:    rev.GameID.String(),
			Rating:    int32(rev.Rating),
			Text:      rev.Text,
			CreatedAt: timestamppb.New(rev.CreatedAt),
			UpdatedAt: timestamppb.New(rev.UpdatedAt),
		})
	}

	return &socialpb.GetUserReviewsResponse{Reviews: revpb}, nil
}

func (h *ReviewHandler) GetGameReviews(ctx context.Context, req *socialpb.GetGameReviewsRequest) (*socialpb.GetGameReviewsResponse, error) {
	log.Info().Str("game_id", req.GameId).Msg("ReviewHandler.GetGameReviews: fetching reviews")

	_, err := microservice.GamesClient.GetGame(ctx, &gamepb.GetGameRequest{
		IdType: &gamepb.GetGameRequest_GameId{
			GameId: req.GameId,
		},
	})
	if err != nil {
		log.Warn().
			Err(err).
			Str("game_id", req.GameId).
			Msg("ReviewHandler.GetGameReviews: game check failed (games-service)")
		return nil, status.Error(codes.NotFound, "game not found")
	}

	reviews, err := h.service.GetReviewsByGame(ctx, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("game_id", req.GameId).
			Msg("ReviewHandler.GetGameReviews: service error")
		return nil, status.Error(codes.Internal, "failed to get reviews")
	}

	revpb := make([]*socialpb.Review, 0, len(reviews))
	for _, rev := range reviews {
		revpb = append(revpb, &socialpb.Review{
			Id:        rev.Id.String(),
			UserId:    rev.UserID.String(),
			GameId:    rev.GameID.String(),
			Rating:    int32(rev.Rating),
			Text:      rev.Text,
			CreatedAt: timestamppb.New(rev.CreatedAt),
			UpdatedAt: timestamppb.New(rev.UpdatedAt),
		})
	}

	return &socialpb.GetGameReviewsResponse{Reviews: revpb}, nil
}
