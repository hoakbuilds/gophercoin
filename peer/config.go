package peer

import "github.com/murlokito/gophercoin/log"

// Config holds the config necessary for the API Server
type Config struct {
	Port     string
	LogLevel log.Level
}
