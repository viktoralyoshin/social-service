package service

import (
	"context"
	"social-service/internal/model"
	"social-service/internal/storage"

	socialpb "github.com/viktoralyoshin/playhub-proto/gen/go/social"
)

type ReviewService struct {
	repo *storage.ReviewRepo
}

func NewReviewService(repo *storage.ReviewRepo) *ReviewService {
	return &ReviewService{
		repo: repo,
	}
}

func (s *ReviewService) CreateReview(ctx context.Context, req *socialpb.CreateReviewRequest) (*model.Review, error) {
	return s.repo.CreateReview(ctx, req)
}

func (s *ReviewService) GetReviewsByGame(ctx context.Context, req *socialpb.GetGameReviewsRequest) ([]*model.Review, error) {
	if req.Limit < 0 {
		req.Limit = 0
	}

	if req.Offset < 0 {
		req.Offset = 0
	}

	return s.repo.GetReviewsByGame(ctx, req)
}
