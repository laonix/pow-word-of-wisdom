package config

import "time"

// ServerParameters holds server settings.
type ServerParameters struct {
	LoggingLevel string `env:"LOGGING_LEVEL" envDefault:"DEBUG"`
	TCPAddr      string `env:"TCP_ADDR" envDefault:":80"`

	Complexity int           `env:"COMPLEXITY" envDefault:"30"`
	WaitPOW    time.Duration `env:"WAIT_POW" envDefault:"1m"`
}
