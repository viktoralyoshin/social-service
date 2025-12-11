package storage

import (
	"context"
	"database/sql"
	"errors"
	"social-service/internal/model"

	"github.com/lib/pq"
	"github.com/rs/zerolog/log"

	socialpb "github.com/viktoralyoshin/playhub-proto/gen/go/social"
	"github.com/viktoralyoshin/utils/pkg/errs"
)

type ReviewRepo struct {
	db *sql.DB
}

func NewReviewRepo(db *sql.DB) *ReviewRepo {
	return &ReviewRepo{
		db: db,
	}
}

func (r *ReviewRepo) CreateReview(ctx context.Context, req *socialpb.CreateReviewRequest) (*model.Review, error) {

	createdReview := &model.Review{}

	query := `
		INSERT INTO social.reviews (user_id, game_id, rating, text)
		VALUES	($1, $2, $3, $4)
		RETURNING id, user_id, game_id, rating, text, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query, req.UserId, req.GameId, req.Rating, req.Text).Scan(
		&createdReview.Id, &createdReview.UserID,
		&createdReview.GameID, &createdReview.Rating,
		&createdReview.Text, &createdReview.CreatedAt,
		&createdReview.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return nil, errs.ErrReviewExists
			}
		}

		return nil, err
	}

	return createdReview, nil
}

func (r *ReviewRepo) GetReviewsByGame(ctx context.Context, req *socialpb.GetGameReviewsRequest) ([]*model.Review, error) {
	reviews := make([]*model.Review, 0, req.Limit)

	query := `
		SELECT id, user_id, game_id, rating, text, created_at, updated_at
		FROM social.reviews
		WHERE game_id = $1
		ORDER BY created_at DESC
		LIMIT = $2 OFFSET = $3
	`

	rows, err := r.db.QueryContext(ctx, query, req.GameId, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("review_repo: failed to close rows")
		}
	}()

	for rows.Next() {
		review := &model.Review{}

		err := rows.Scan(
			&review.Id,
			&review.UserID,
			&review.GameID,
			&review.Rating,
			&review.Text,
			&review.CreatedAt,
			&review.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}
