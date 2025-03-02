package postgres

import (
	"auth/internal/config"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewDatabase(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	connInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	client, err := sqlx.Connect("postgres", connInfo)
	if err != nil {
		return nil, err
	}

	err = client.Ping()
	if err != nil {
		return nil, err
	}

	return client, nil
}
