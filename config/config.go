package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type ConfigList struct {
	ApiKey    string
	ApiSecret string
}

var Config ConfigList

func LoadConfig() *ConfigList {

	err := godotenv.Load()
	if err != nil {
		log.Printf(".envファイルが見つかりません。: %v", err)
	}

	return &ConfigList{
		ApiKey:    os.Getenv("API_KEY"),
		ApiSecret: os.Getenv("API_SECRET"),
	}
}
