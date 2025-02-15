package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DB_URL          string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	configFilePath, err := getConfigFilePath()

	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(configFilePath)

	if err != nil {
		return Config{}, err
	}

	var config Config

	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username

	err := write(*c)

	if err != nil {
		return err
	}

	return nil
}

func write(cfg Config) error {
	configFilePath, err := getConfigFilePath()

	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(configFilePath, bytes, 0666)

	if err != nil {
		return err
	}

	return nil
}

func getConfigFilePath() (string, error) {
	homedir, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	return homedir + "/.gatorconfig.json", nil
}
