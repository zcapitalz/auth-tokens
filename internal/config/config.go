package config

import (
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        Env              `env:"ENV" env-required:"true"`
	LogLevel   string           `env:"LOG_LEVEL" env-default:"warn"`
	DBConfig   DatabaseConfig   `env-prefix:"DB_"`
	HTTPServer HTTPServerConfig `env-prefix:"HTTP_SERVER_"`
	SMTPServer SMTPServerConfig `env-prefix:"SMTP_"`
	Auth       AuthConfig       `env-prefix:"AUTH_"`
	Emails     Emails           `env-prefix:"EMAILS_"`
}

type Env string

const (
	EnvLocal Env = "local"
	EnvDev   Env = "dev"
	EnvProd  Env = "prod"
)

type HTTPServerConfig struct {
	Host string `env:"HOST" env-required:"true"`
	Port string `env:"PORT" env-required:"true"`
}

type DatabaseConfig struct {
	Host     string `env:"HOST" env-required:"true"`
	Port     string `env:"PORT" env-required:"true"`
	DBName   string `env:"NAME" env-required:"true"`
	SSLMode  string `env:"SSL_MODE" env-required:"true"`
	Username string `env:"USERNAME" env-required:"true"`
	Password string `env:"PASSWORD" env-required:"true"`
}

type AuthConfig struct {
	AccessTokenDuration  time.Duration `env:"ACCESS_TOKEN_DURATION" env-required:"true"`
	RefreshTokenDuration time.Duration `env:"REFRESH_TOKEN_DURATION" env-required:"true"`
	JWTPrivateKey        string        `env:"JWT_PRIVATE_KEY" env-required:"true"`
}

type SMTPServerConfig struct {
	Host     string `env:"HOST" env-required:"true"`
	Port     int    `env:"PORT" env-required:"true"`
	Username string `env:"USERNAME" env-required:"true"`
	Password string `env:"PASSWORD" env-required:"true"`
}

type Emails struct {
	SupportEmail string `env:"SUPPORT_EMAIL" env-required:"true"`
}

var (
	once sync.Once
	cfg  Config
)

func New() (Config, error) {
	var err error
	once.Do(func() {
		err = cleanenv.ReadEnv(&cfg)
	})

	return cfg, err
}
