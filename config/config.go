package config

type ConfigList struct {
	ApiKey    string
	ApiSecret string
	LogFile   string
}

var Config ConfigList

func LoadConfig() (*ConfigList, error) {

	envValues, err := LoadEnValues()
	if err != nil {
		return nil, err
	}

	iniValues, err := LoadIniValues("config.ini")
	if err != nil {
		return nil, err
	}

	cfg := &ConfigList{
		ApiKey:    envValues["ApiKey"],
		ApiSecret: envValues["ApiSecret"],
		LogFile:   iniValues["log_file"],
	}

	return cfg, nil
}
