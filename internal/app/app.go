package app

import (
	"net"
	"social-service/internal/config"
	"social-service/internal/database"
	"social-service/internal/grpc"

	"github.com/rs/zerolog/log"
)

func Start(cfg *config.Config) {
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to connect to database")
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close database connection")
		}
	}()

	if err := database.Migrate(db); err != nil {
		log.Fatal().Err(err).Msg("migration failed")
	}

	addr := ":" + cfg.GRPCPort
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Err(err).Str("addr", addr).Msg("failed to listen tcp")
	}

	s := grpc.Init(db)

	log.Info().
		Str("port", cfg.GRPCPort).
		Str("service", "social-service").
		Msg("gRPC server started")

	if err := s.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("gRPC server stopped unexpectedly")
	}
}
