package main

import (
	"context"
	"fmt"
	"gotrading/bitflyer"
	"gotrading/config"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("設定読み込みに失敗しました: %v", err)
	}

	apiClient := bitflyer.New(cfg.ApiKey, cfg.ApiSecret)

	//debug: APIキーとシークレットが正しく渡されているかの確認
	fmt.Printf("Key: %s, Secret: %s\n", cfg.ApiKey, cfg.ApiSecret)

	// fmt.Println(apiClient.GetBalance())
	// balance, err := apiClient.GetBalance()
	// if err != nil {
	// 	log.Fatalf("Error: %v", err)
	// }

	// fmt.Printf("Balance: %+v\n", balance)

	// ticker, _ := apiClient.GetTicker("BTC_USD")
	// fmt.Println(ticker.GetMidPrice())
	// fmt.Println(ticker.DateTime())
	// fmt.Println(ticker.TruncateDateTime(time.Hour))

	// WebSocket用のティッカーチャンネル
	tickerChannel := make(chan bitflyer.Ticker)

	// 終了シグナルを捕捉するためのコンテキストを作成
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 終了シグナルをキャッチするゴルーチン
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		<-signalChan
		fmt.Println("終了シグナルを受信しました。シャットダウンします...")
		cancel()
	}()

	// WebSocketを使用したリアルタイムデータの受信
	go func() {
		apiClient.GetRealTimeTicker(ctx, "BTC_JPY", tickerChannel)
	}()

	// ティッカー情報を処理するゴルーチン
	go func() {
		for ticker := range tickerChannel {
			fmt.Printf("Received Ticker: %+v\n", ticker)
		}
	}()

	// メインスレッドをブロック
	<-ctx.Done()
	fmt.Println("コンテキストがキャンセルされました。プログラムを終了します。")
	time.Sleep(1 * time.Second) // リソース解放のための待機
}
