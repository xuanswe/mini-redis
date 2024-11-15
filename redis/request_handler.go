package redis

import (
	"github.com/rs/zerolog/log"
	respModels "github.com/xuanswe/mini-redis/internal/models"
)

func handleRequest(request *respModels.Request) ([]byte, error) {
	log.Debug().Msgf("Processing request: %v", request)
	return []byte("Hello " + request.Data + "!"), nil
}
