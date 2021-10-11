package utils

import (
	"github.com/jmoiron/sqlx"
)

// DBConfig db config
type DBConfig struct {
	Driver       string `mapstructure:"driver"`
	Source       string `mapstructure:"source"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

// Open is a sqlx.Open wraper.
func Open(config *DBConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open(config.Driver, config.Source)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	return db, nil
}
