package config

import "os"

type Config struct {
	Addr      string
	Secret    string
	UserSvc   string
	OrdersSvc string
}

func NewConfig() *Config {
	return &Config{
		Addr:      getenv("GATEWAY_ADDR", ":8090"),
		Secret:    must("JWT_SECRET"),
		UserSvc:   must("USER_SERVICE_URL"),
		OrdersSvc: must("ORDERS_SERVICE_URL"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func must(k string) string {
	v := os.Getenv(k)
	if v == "" {
		panic("missing env: " + k)
	}
	return v
}
