package config

import (
	"os"
	"strconv"
)

type HTTPConfig struct {
	Addr string
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type LoggerConfig struct {
	Level string
}

type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
	Logger   LoggerConfig
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return def
}

func Load() Config {
	return Config{
		HTTP: HTTPConfig{
			Addr: getenv("HTTP_ADDR", ":8082"),
		},
		Postgres: PostgresConfig{
			Host:     getenv("POSTGRES_HOST", "localhost"),
			Port:     getenvInt("POSTGRES_PORT", 5432),
			User:     getenv("POSTGRES_USER", "orders_user"),
			Password: getenv("POSTGRES_PASSWORD", "orders_pass"),
			DBName:   getenv("POSTGRES_DB", "orders_db"),
		},
		Logger: LoggerConfig{
			Level: getenv("LOG_LEVEL", "info"),
		},
	}
}

func (p PostgresConfig) DSN() string {
	return "postgres://" + p.User + ":" + p.Password +
		"@" + p.Host + ":" + strconv.Itoa(p.Port) +
		"/" + p.DBName
}
