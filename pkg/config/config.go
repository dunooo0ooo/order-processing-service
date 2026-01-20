package config

import (
	"os"
	"strconv"
	"strings"
)

type HTTPConfig struct {
	Addr string
}

type PostgresConfig struct {
	Host          string
	Port          int
	User          string
	Password      string
	DBName        string
	MigrationsDir string
}

type LoggerConfig struct {
	Level string
}

type KafkaConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

type Config struct {
	HTTP     HTTPConfig
	Postgres PostgresConfig
	Logger   LoggerConfig
	Kafka    KafkaConsumerConfig
	Cache    CacheConfig
}

type CacheConfig struct {
	Limit int
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
			Addr: getenv("HTTP_ADDR", ":8081"),
		},
		Postgres: PostgresConfig{
			Host:          getenv("POSTGRES_HOST", "localhost"),
			Port:          getenvInt("POSTGRES_PORT", 5432),
			User:          getenv("POSTGRES_USER", "orders_user"),
			Password:      getenv("POSTGRES_PASSWORD", "orders_pass"),
			DBName:        getenv("POSTGRES_DB", "orders_db"),
			MigrationsDir: getenv("MIGRATIONS_DIR", "./migrations"),
		},
		Logger: LoggerConfig{
			Level: getenv("LOG_LEVEL", "info"),
		},
		Kafka: KafkaConsumerConfig{
			Brokers: strings.Split(getenv("KAFKA_BROKERS", "localhost:29092"), ","),
			Topic:   getenv("KAFKA_TOPIC", "orders"),
			GroupID: getenv("KAFKA_GROUP_ID", "order-information-service"),
		},
		Cache: CacheConfig{
			Limit: getenvInt("CACHE_LIMIT", 500),
		},
	}
}

func (p PostgresConfig) DSN() string {
	return "postgres://" + p.User + ":" + p.Password +
		"@" + p.Host + ":" + strconv.Itoa(p.Port) +
		"/" + p.DBName
}
