package redis

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/xuanswe/mini-redis/internal/resp/consts/commands"
	respEncoders "github.com/xuanswe/mini-redis/internal/resp/encoders"
	respModels "github.com/xuanswe/mini-redis/internal/resp/models"
	"strconv"
	"strings"
	"time"
)

func handleRequest(request *respModels.Request) ([]byte, error) {
	log.Debug().Msgf("Processing request: %v", request)

	switch strings.ToUpper(request.Command) {
	case commands.Ping:
		return handlePingCommand(request.Args)
	case commands.Echo:
		return handleEchoCommand(request.Args)
	case commands.Set:
		return handleSetCommand(request.Args)
	case commands.Get:
		return handleGetCommand(request.Args)
	}

	return []byte(respEncoders.SimpleError(fmt.Sprintf("ERR unsupported command %s", request.Command))), nil
}

// https://redis.io/docs/latest/commands/ping
// PING [message]
func handlePingCommand(args []string) ([]byte, error) {
	validator := func(len int) bool {
		return len <= 1
	}
	if message, ok := validateArgsLength(commands.Ping, args, validator); !ok {
		return []byte(respEncoders.SimpleError(message)), nil
	}

	str := "PONG"
	if len(args) != 0 {
		str = args[0]
	}
	return []byte(respEncoders.SimpleString(str)), nil
}

// https://redis.io/docs/latest/commands/echo
// ECHO message
func handleEchoCommand(args []string) ([]byte, error) {
	validator := func(len int) bool {
		return len == 1
	}
	if message, ok := validateArgsLength(commands.Echo, args, validator); !ok {
		return []byte(respEncoders.SimpleError(message)), nil
	}

	return []byte(respEncoders.BulkString(args[0])), nil
}

type storeValue struct {
	value     string
	expiredAt time.Time
}

// TODO: handling race conditions when accessing the same key concurrently
var kvStore = map[string]storeValue{}

// https://redis.io/docs/latest/commands/set
// Only handle simple cases: SET key value [PX milliseconds]
func handleSetCommand(args []string) ([]byte, error) {
	validator := func(len int) bool {
		return len == 2 || len == 4
	}
	if message, ok := validateArgsLength(commands.Set, args, validator); !ok {
		return []byte(respEncoders.SimpleError(message)), nil
	}

	key := args[0]
	value := args[1]
	var expiredAt time.Time

	if len(args) == 4 {
		pxArg := strings.ToUpper(args[2])
		if pxArg != "PX" {
			return []byte(respEncoders.SimpleError("ERR syntax error: expect PX option at 3rd argument")), nil
		}
		px, err := strconv.Atoi(args[3])
		if err != nil || px <= 0 {
			return []byte(respEncoders.SimpleError("ERR syntax error: PX's value should be a positive number in milliseconds")), nil
		}

		expiredAt = time.Now().Add(time.Duration(px) * time.Millisecond)
	}

	kvStore[key] = storeValue{value: value, expiredAt: expiredAt}
	return []byte(respEncoders.SimpleString("OK")), nil
}

// https://redis.io/docs/latest/commands/get
// GET key
func handleGetCommand(args []string) ([]byte, error) {
	validator := func(len int) bool {
		return len == 1
	}
	if message, ok := validateArgsLength(commands.Get, args, validator); !ok {
		return []byte(respEncoders.SimpleError(message)), nil
	}

	key := args[0]
	if storedValue, ok := kvStore[key]; ok {
		if !storedValue.expiredAt.IsZero() && time.Now().After(storedValue.expiredAt) {
			delete(kvStore, key)
			return []byte(respEncoders.NullBulkString), nil
		}

		return []byte(respEncoders.BulkString(storedValue.value)), nil
	}
	return []byte(respEncoders.NullBulkString), nil
}

func validateArgsLength(name commands.CommandType, args []string, validator func(len int) bool) (string, bool) {
	if !validator(len(args)) {
		return fmt.Sprintf("ERR wrong number of arguments for '%s' command", name), false
	}

	return "", true
}
