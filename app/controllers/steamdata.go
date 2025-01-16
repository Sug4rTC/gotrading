package controllers

import (
	"context"
	"gotrading/app/models"
	"gotrading/bitflyer"
	"gotrading/config"
	"log"
)

func StreamIngestionData() {

	ctx, _ := context.WithCancel(context.Background())

	tickerChannel := make(chan bitflyer.Ticker)
	apiClient := bitflyer.New(config.Config.ApiKey, config.Config.ApiSecret)

	go apiClient.GetRealTimeTicker(ctx, config.Config.ProductCode, tickerChannel)
	log.Printf("ProductCode in config: %s", config.Config.ProductCode)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Context canceled, stopping StreamIngestionData.")
				return
			case ticker, ok := <-tickerChannel:
				if !ok {
					log.Println("Ticker channel closed, exiting goroutine.")
					return
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
	}()
}
