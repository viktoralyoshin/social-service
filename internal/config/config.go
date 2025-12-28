package config

import "os"

type Config struct {
	DBUser          string
	DBHost          string
	DBPassword      string
	DBPort          string
	DBName          string
	GRPCPort        string
	GameServiceAddr string
	AuthServiceAddr string
	KafkaAddr       string
	Env             string
}

func Load() *Config {
	return &Config{
		DBUser:          os.Getenv("DB_USER"),
		DBName:          os.Getenv("DB_NAME"),
		DBHost:          os.Getenv("DB_HOST"),
		DBPassword:      os.Getenv("DB_PASSWORD"),
		DBPort:          os.Getenv("DB_PORT"),
		GRPCPort:        os.Getenv("GRPC_PORT"),
		GameServiceAddr: os.Getenv("GAME_SERVICE_ADDR"),
		AuthServiceAddr: os.Getenv("AUTH_SERVICE_ADDR"),
		Env:             os.Getenv("ENV"),
		KafkaAddr:       os.Getenv("KAFKA_ADDR"),
	}
}
