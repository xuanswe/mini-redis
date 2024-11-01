package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xuanswe/mini-redis/internal/server"
	"os"
	"time"
)

func main() {
	setupLogger()

	redisConfig := server.RedisServerConfig{
		Host: "0.0.0.0",
		Port: "6379",
	}
	redisServer, err := server.NewRedisServer(redisConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating Redis server")
	}

	if err := redisServer.Start(); err != nil {
		log.Fatal().Err(err).Msg("Error starting Redis server")
	}

	defer func() {
		if err := redisServer.Close(); err != nil {
			log.Fatal().Err(err).Msg("Error closing Redis server")
		}
	}()
}

func setupLogger() {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Set the output to console
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339Nano,
	})
}
