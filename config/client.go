package config

// ClientParameters holds a client settings.
type ClientParameters struct {
	LoggingLevel string `env:"LOGGING_LEVEL" envDefault:"DEBUG"`
	ServerAddr   string `env:"SERVER_ADDR" envDefault:":80"`
}
