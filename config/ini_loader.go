package config

import (
	"log"
	"os"

	"gopkg.in/ini.v1"
)

func LoadIniValues(filepath string) (map[string]string, error) {
	cfg, err := ini.Load(filepath)
	if err != nil {
		log.Printf("Failed to read file: %v ", err)
		os.Exit(1)
	}

	return map[string]string{
		"log_file":       cfg.Section("gotrading").Key("log_file").String(),
		"product_code":   cfg.Section("gotrading").Key("product_code").String(),
		"trade_duration": cfg.Section("gotrading").Key("trade_duration").String(),
	}, nil
}
