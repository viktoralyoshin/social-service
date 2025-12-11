package handlers

import (
	"context"
	"errors"
	"social-service/internal/service"
	"social-service/internal/utils"

	socialpb "github.com/viktoralyoshin/playhub-proto/gen/go/social"
	"github.com/viktoralyoshin/utils/pkg/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ReviewHandler struct {
	socialpb.UnimplementedSocialServiceServer
	service *service.ReviewService
}

func NewReviewHandler(service *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{
		service: service,
	}
}

func (h *ReviewHandler) CreateReview(ctx context.Context, req *socialpb.CreateReviewRequest) (*socialpb.CreateReviewResponse, error) {
	userId, err := utils.GetUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	req.UserId = userId

	review, err := h.service.CreateReview(ctx, req)
	if err != nil {
		if errors.Is(err, errs.ErrReviewExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		return nil, status.Error(codes.Internal, "internal error during review creation")
	}

	return &socialpb.CreateReviewResponse{
		Id:        review.Id.String(),
		UserId:    review.UserID.String(),
		GameId:    review.GameID.String(),
		Rating:    int32(review.Rating),
		Text:      review.Text,
		CreatedAt: timestamppb.New(review.CreatedAt),
		UpdatedAt: timestamppb.New(review.CreatedAt),
	}, nil
}

func (h *ReviewHandler) GetGameReviews(ctx context.Context, req *socialpb.GetGameReviewsRequest) (*socialpb.GetGameReviewsResponse, error) {

	//НУЖНО НЕ ЗАБЫТЬ ОПТРАВИТЬ ЗАПРОС НА GAMESERVICE ЧТОБЫ ПРОВЕРИТЬ ЕСЛТЬ ЛИ ИГРА ВООБЩЕ ТАКАЯ ИЛИ НЕТ GetGame

	reviews, err := h.service.GetReviewsByGame(ctx, req)
	if err != nil {
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

	return &socialpb.GetGameReviewsResponse{
		Reviews: revpb,
	}, nil
}
