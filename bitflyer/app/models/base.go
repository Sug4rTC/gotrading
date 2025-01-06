package models

import (
	"database/sql"
	"fmt"
	"gotrading/config"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	tableNameSignalEvents = "signal_events"
)

var DbConnection *sql.DB

func GetCandleTableName(productCode string, duration time.Duration) string {
	return fmt.Sprintf("%s_%s", productCode, duration)
}

func InitializeDatabase() error {

	var err error

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("設定読み込みに失敗しました: %v", err)
	}

	DbConnection, err = sql.Open(cfg.SQLDriver, cfg.DBName)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	cmd := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
		time DATETIME PRIMARY KEY NOT NULL,
		product_code STRING,
		side STRING,
		price FLOAT,
		size FLOAT)`, tableNameSignalEvents)
	if _, err := DbConnection.Exec(cmd); err != nil {
		return fmt.Errorf("failed to create DataBase: %w", err)
	}

	for _, duration := range cfg.Durations {
		tableName := GetCandleTableName(cfg.ProductCode, duration)
		log.Printf("Creating table: %s", tableName)
		c := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
			time DATETIME PRIMARY KEY NOT NULL,
			open FLOAT,
			close FLOAT,
			high FLOAT,
			low FLOAT,
			volume FLOAT)`, tableName)
		if _, err := DbConnection.Exec(c); err != nil {
			return fmt.Errorf("failed to create candle table: %w", err)
		}
	}

	return nil
}
