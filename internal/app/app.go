package app

import (
	"net"
	"social-service/internal/config"
	"social-service/internal/database"
	"social-service/internal/grpc"
	"social-service/internal/microservice"
	"social-service/internal/producer"

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

	producer := producer.NewRatingProducer(cfg.KafkaAddr, "review_events")
	defer func() {
		if err := producer.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close kafka producer")
		}
	}()

	s := grpc.Init(db, producer)

	microservice.Connect(cfg)

	log.Info().
		Str("port", cfg.GRPCPort).
		Str("service", "social-service").
		Msg("gRPC server started")

	if err := s.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("gRPC server stopped unexpectedly")
	}
}
