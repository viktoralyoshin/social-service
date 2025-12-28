package microservice

import (
	"social-service/internal/config"

	"github.com/rs/zerolog/log"
	authpb "github.com/viktoralyoshin/playhub-proto/gen/go/auth"
	gamepb "github.com/viktoralyoshin/playhub-proto/gen/go/games"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var GamesClient gamepb.GameServiceClient
var AuthClient authpb.AuthServiceClient

func Connect(cfg *config.Config) {
	gamesServiceConn := connect(cfg.GameServiceAddr)
	authServiceConn := connect(cfg.AuthServiceAddr)

	AuthClient = authpb.NewAuthServiceClient(authServiceConn)
	GamesClient = gamepb.NewGameServiceClient(gamesServiceConn)
}

func connect(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().
			Err(err).
			Str("target_addr", addr).
			Msg("failed to initialize grpc connection")
	}

	return conn
}
