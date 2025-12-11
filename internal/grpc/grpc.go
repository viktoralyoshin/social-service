package grpc

import (
	"database/sql"
	"social-service/internal/handlers"
	"social-service/internal/service"
	"social-service/internal/storage"

	socialpb "github.com/viktoralyoshin/playhub-proto/gen/go/social"
	"google.golang.org/grpc"
)

func Init(db *sql.DB) *grpc.Server {
	s := grpc.NewServer()

	socialRepo := storage.NewReviewRepo(db)
	socialService := service.NewReviewService(socialRepo)
	socialHandler := handlers.NewReviewHandler(socialService)

	socialpb.RegisterSocialServiceServer(s, socialHandler)

	return s
}
