package server

import (
	"github.com/rs/zerolog/log"
	"net"
)

type RedisServerInterface interface {
	Start() error
	Close() error
}

type RedisServer struct {
	config RedisServerConfig
}

type RedisServerConfig struct {
	Host string
	Port string
}

func NewRedisServer(config RedisServerConfig) (RedisServerInterface, error) {
	return &RedisServer{config}, nil
}

func (k *RedisServer) Start() error {
	listener, err := net.Listen("tcp", k.config.Host+":"+k.config.Port)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to bind to %s:%s", k.config.Host, k.config.Port)
		return err
	}
	defer func() {
		err := listener.Close()
		if err != nil {
			log.Error().Err(err).Msg("Error closing listener")
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("Error accepting connection")
			continue
		}

		go func() {
			err := handleConnection(conn)
			if err != nil {
				log.Error().Err(err).Msg("Error handling connection")
			}
		}()
	}
}

func (k *RedisServer) Close() error {
	// TODO: closing listener, connections and other resources gracefully
	return nil
}

func handleConnection(conn net.Conn) error {
	// TODO
	return nil
}
