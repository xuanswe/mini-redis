package models

import (
	"github.com/xuanswe/mini-redis/internal/resp/consts/commands"
)

type Request struct {
	RemoteAddr string
	Command    commands.CommandType
	Args       []string
}
