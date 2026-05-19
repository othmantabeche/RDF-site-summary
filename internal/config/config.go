package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Db_url            string `json:"Db_url"`
	Current_user_name string `json:"Current_user_name"`
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Error HOME not found")
	}

	path := fmt.Sprintf("%s/%s", homePath, configFileName)
	return path, nil
}

func Read() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, fmt.Errorf("Error traing to find path: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("Error traing to read the file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("Error: %v", err)
	}

	return config, nil
}

func (cfg *Config) SetUser(name string) error {
	cfg.Current_user_name = name

	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	dataByte, err := json.MarshalIndent(cfg, "", "")
	if err != nil {
		return fmt.Errorf("Error traying to marshal the content: %v", err)
	}

	if err := os.WriteFile(path, dataByte, 0666); err != nil {
		return fmt.Errorf("Error traying to write: %v", err)
	}

	return nil
}
