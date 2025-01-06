package controllers

import (
	"context"
	"gotrading/bitflyer"
	"gotrading/bitflyer/app/models"
	"gotrading/config"
	"log"
)

func StreamIngestionData() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tickerChannel := make(chan bitflyer.Ticker)

	apiClient := bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)
	go apiClient.GetRealTimeTicker(ctx, config.Config.ProductCode, tickerChannel)
	log.Printf("ProductCode in config: %s", config.Config.ProductCode)
	for ticker := range tickerChannel {

		if ticker.ProductCode == "" {
			ticker.ProductCode = config.Config.ProductCode
			log.Printf("Warning: Ticker ProductCode was empty. Fallback to config: %s", ticker.ProductCode)
		}

		log.Printf("action=StreamIngestionData, %v", ticker)
		for _, duration := range config.Config.Durations {
			log.Printf("Calling CreateCandleWithDuration for ProductCode: %s, Duration: %v", ticker.ProductCode, duration)
			isCreated := models.CreateCandleWithDuration(ticker, ticker.ProductCode, duration)

			if isCreated {
				log.Printf("New candle created for ProductCode: %s, Duration: %v", ticker.ProductCode, duration)
			}
		}
	}
}
