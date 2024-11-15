package redis

import (
	"github.com/rs/zerolog/log"
	"github.com/xuanswe/mini-redis/internal/models"
)

func handleRequest(request *models.Request) ([]byte, error) {
	log.Debug().Msgf("Processing request: %v", request)
	return []byte("Hello " + request.Data + "!"), nil
}
