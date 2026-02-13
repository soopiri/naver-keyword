package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultConfigPath = "config.json"
)

var defaultConfig = Config{
	Naver: NaverConfig{
		ClientID:     "",
		ClientSecret: "",
	},
	CreatorAdvisor: CreatorAdvisorConfig{
		LoginID:  "",
		Password: "",
	},
	SearchAd: SearchAdConfig{
		BaseURL:    "https://api.searchad.naver.com",
		CustomerID: "",
		AccessKey:  "",
		SecretKey:  "",
	},
	Scoring: Scoring{
		WPC:   0.4,
		WMob:  0.6,
		K:     1.0,
		Alpha: 0.5,
		T:     0.2,
		R:     0.3,
	},
	Seeds: []string{},
	Usage: Usage{
		Date: "",
		Used: 0,
	},
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return defaultConfigPath, nil
	}
	configDir := filepath.Join(homeDir, ".pick-keyword")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return defaultConfigPath, nil
	}
	return filepath.Join(configDir, defaultConfigPath), nil
}

func getTodayKey() string {
	now := time.Now()
	return now.Format("2006-01-02")
}

func normalizeUsage(usage Usage) Usage {
	today := getTodayKey()
	if usage.Date != today {
		return Usage{
			Date: today,
			Used: 0,
		}
	}
	return usage
}

func readConfig() (Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return defaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			config := defaultConfig
			config.Usage = normalizeUsage(config.Usage)
			return config, nil
		}
		return defaultConfig, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return defaultConfig, err
	}

	config.Usage = normalizeUsage(config.Usage)
	return config, nil
}

func writeConfig(config Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	config.Usage = normalizeUsage(config.Usage)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func incrementUsage(delta int) (Usage, error) {
	config, err := readConfig()
	if err != nil {
		return Usage{}, err
	}

	usage := normalizeUsage(config.Usage)
	if delta < 0 {
		delta = 0
	}
	usage.Used += delta

	config.Usage = usage
	if err := writeConfig(config); err != nil {
		return usage, err
	}

	return usage, nil
}
