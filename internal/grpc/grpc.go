package grpc

import (
	"database/sql"
	"social-service/internal/handlers"
	"social-service/internal/producer"
	"social-service/internal/service"
	"social-service/internal/storage"

	gamepb "github.com/viktoralyoshin/playhub-proto/gen/go/games"
	socialpb "github.com/viktoralyoshin/playhub-proto/gen/go/social"
	"google.golang.org/grpc"
)

var GamesClient gamepb.GameServiceClient

func Init(db *sql.DB, producer *producer.RatingProducer) *grpc.Server {
	s := grpc.NewServer()

	socialRepo := storage.NewReviewRepo(db)
	socialService := service.NewReviewService(socialRepo)
	socialHandler := handlers.NewReviewHandler(socialService, producer)

	socialpb.RegisterSocialServiceServer(s, socialHandler)

	return s
}
