package config

import "os"

type Config struct {
	Addr              string
	SessionCookieName string
}

func Load() Config {
	addr := os.Getenv("APP_ADDR")
	if addr == "" {
		addr = ":8090"
	}

	cookieName := os.Getenv("SESSION_COOKIE_NAME")
	if cookieName == "" {
		cookieName = "sid"
	}

	return Config{
		Addr:              addr,
		SessionCookieName: cookieName,
	}
}
