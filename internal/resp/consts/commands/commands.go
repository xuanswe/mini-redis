package commands

type CommandType = string

const (
	Ping CommandType = "PING"
	Echo CommandType = "ECHO"
	Set  CommandType = "SET"
	Get  CommandType = "GET"
)
