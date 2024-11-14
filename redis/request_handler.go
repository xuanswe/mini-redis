package redis

import (
	"github.com/rs/zerolog/log"
	respModels "github.com/xuanswe/mini-redis/internal/resp/models"
)

func handleRequest(request *respModels.Request) ([]byte, error) {
	log.Debug().Msgf("Processing request: %v", request)
	return make([]byte, 0), nil
}
