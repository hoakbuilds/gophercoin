package api

import "github.com/murlokito/gophercoin/log"

// Config holds the config necessary for the API Server
type Config struct {
	Port      string
	Protected bool
	Password  string
	LogLevel  log.Level
	Routes    []Route
}
