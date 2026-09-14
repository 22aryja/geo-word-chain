package config

import (
	"errors"
	"io/fs"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Config(key string) string {
	if err := godotenv.Load(".env"); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Printf("loading .env: %v", err)
	}
	return os.Getenv(key)
}
