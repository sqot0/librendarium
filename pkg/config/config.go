package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config contains the credentials needed to talk to Librus and Google Calendar.
type Config struct {
	LibrusLogin               string
	LibrusPassword            string
	GoogleCalendarClientEmail string
	GoogleCalendarPrivateKey  string
	GoogleCalendarID          string
	SyncInterval              string
}

// Load reads environment variables (optionally from a .env file) and validates them.
func Load(envPath ...string) (Config, error) {
	_ = godotenv.Load(envPath...)

	get := func(key string) (string, error) {
		if value := os.Getenv(key); value != "" {
			return value, nil
		}
		return "", fmt.Errorf("missing required environment variable %s", key)
	}

	login, err := get("LIBRUS_LOGIN")
	if err != nil {
		return Config{}, err
	}

	password, err := get("LIBRUS_PASSWORD")
	if err != nil {
		return Config{}, err
	}

	email, err := get("GOOGLE_CALENDAR_CLIENT_EMAIL")
	if err != nil {
		return Config{}, err
	}

	privateKey, err := get("GOOGLE_CALENDAR_PRIVATE_KEY")
	if err != nil {
		return Config{}, err
	}

	id, err := get("GOOGLE_CALENDAR_ID")
	if err != nil {
		return Config{}, err
	}

	return Config{
		LibrusLogin:               login,
		LibrusPassword:            password,
		GoogleCalendarClientEmail: email,
		GoogleCalendarPrivateKey:  privateKey,
		GoogleCalendarID:          id,
		SyncInterval:              os.Getenv("SYNC_INTERVAL"),
	}, nil
}

// PrivateKey returns the service account PEM key with literal `\\n` replaced by actual newlines.
func (c Config) PrivateKey() string {
	return strings.ReplaceAll(c.GoogleCalendarPrivateKey, `\\n`, "\n")
}
