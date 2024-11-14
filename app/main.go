package main

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/xuanswe/mini-redis/internal/support"
	"github.com/xuanswe/mini-redis/redis"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	support.SetupLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture OS interrupted signals
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, syscall.SIGINT, syscall.SIGTERM)

	redisConfig := redis.ServerConfig{
		Host:            "0.0.0.0",
		Port:            "6379",
		ConnIdleTimeout: 1 * time.Minute,
	}
	redisServer, err := redis.NewServer(redisConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating Redis server")
	}

	go func() {
		if err := redisServer.Start(); err != nil {
			log.Fatal().Err(err).Msg("Error starting Redis server")
		}

		cancel()
	}()

	// Handle OS interrupted signals in a separated goroutine
	go func() {
		sig := <-interrupts
		log.Info().Msgf("Signal intercepted %v", sig)

		if err := redisServer.Shutdown(); err != nil {
			log.Fatal().Err(err).Msg("Error closing Redis server")
		}

		cancel()
	}()

	// Wait for the shutdown flow to send true
	<-ctx.Done()
}
