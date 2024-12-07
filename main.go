package main

import (
	"fmt"
	"gotrading/config"
)

func main() {
	cfg := config.LoadConfig()

	if cfg.ApiKey == "" {
		fmt.Printf("APIキーが設定されていません。")
		return
	}

	fmt.Printf("ApiKey: %s\n", cfg.ApiKey)

	if cfg.ApiSecret == "" {
		fmt.Printf("APISCRETキーが設定されていません。")
		return
	}

	fmt.Printf("ApiSecret: %s\n", cfg.ApiSecret)
}
