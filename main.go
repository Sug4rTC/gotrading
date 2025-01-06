package main

import (
	"gotrading/bitflyer/app/controllers"
	"gotrading/bitflyer/app/models"
	"gotrading/config"
	"gotrading/utils"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to read configfile: %v", err)
	}
	config.Config = *cfg

	utils.LoggingSetting(cfg.LogFile)

	if err := models.InitializeDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// ティッカーのリアルタイム処理開始
	log.Println("Starting StreamIngestionData...")
	controllers.StreamIngestionData()

}
