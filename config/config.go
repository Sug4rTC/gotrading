package config

import (
	"fmt"
	"log"
	"time"
)

type ConfigList struct {
	ApiKey      string
	ApiSecret   string
	LogFile     string
	ProductCode string

	TradeDuration time.Duration
	Durations     map[string]time.Duration
	DBName        string
	SQLDriver     string
	Port          string
}

var Config ConfigList

func LoadConfig() (*ConfigList, error) {

	envValues, err := LoadEnValues()
	if err != nil {
		return nil, fmt.Errorf("failed to load enviroment values: %w", err)
	}

	iniValues, err := LoadIniValues("config.ini")
	if err != nil {
		return nil, fmt.Errorf("failed to load ini values: %w", err)
	}

	tradeDuration, err := time.ParseDuration(iniValues["trade_duration"])
	if err != nil {
		log.Printf("failed to parse trade_duration:%v", err)
		log.Printf("invailed trade_duration value: %s", iniValues["trade_duration"])
		tradeDuration = time.Minute
	}

	durations := map[string]time.Duration{
		"1s": time.Second,
		"1m": time.Minute,
		"1h": time.Hour,
	}

	cfg := &ConfigList{
		ApiKey:        envValues["ApiKey"],
		ApiSecret:     envValues["ApiSecret"],
		DBName:        envValues["DBName"],
		SQLDriver:     envValues["DRIVER"],
		Port:          envValues["PORT"],
		LogFile:       iniValues["log_file"],
		ProductCode:   iniValues["product_code"],
		TradeDuration: tradeDuration,
		Durations:     durations,
	}

	return cfg, nil

}
