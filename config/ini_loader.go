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
		"log_file": cfg.Section("gotradinf").Key("log_file").String(),
	}, nil
}
