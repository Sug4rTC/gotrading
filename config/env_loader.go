package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnValues() (map[string]string, error) {

	if err := godotenv.Load(); err != nil {
		log.Printf(".envファイルが見つかりません: %v", err)
	}

	return map[string]string{
		"ApiKey":    os.Getenv("API_KEY"),
		"ApiSecret": os.Getenv("API_SECRET"),
	}, nil
}
