package config

import (
	"fmt"
	"os"
	"path/filepath"
	. "q/types"

	_ "embed"

	"gopkg.in/yaml.v2"
)

type AppConfig struct {
	Models      []ModelConfig `yaml:"models"`
	Preferences Preferences   `yaml:"preferences"`
	Version     string        `yaml:"config_format_version"`
}

// //go:embed config.yaml
// var embeddedConfigFile []byte

//go:embed config.yaml
var embeddedConfigFile []byte
var configFilePath string = ".shelly-ai/config.yaml"
var backupConfigFilePath string = ".shelly-ai/.backup-config.yaml"

func FullFilePath(relativeFilePath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %w", err)
	}
	configFilePath := filepath.Join(homeDir, relativeFilePath)
	return configFilePath, nil
}

func LoadAppConfig() (config AppConfig, err error) {
	filePath, err := FullFilePath(configFilePath)
	if err != nil {
		return config, fmt.Errorf("error getting config file path: %w", err)
	}

	// if file doesn't exist, create it with defaults
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return createConfigWithDefaults(filePath)
	}
	return loadExistingConfig(filePath)
}

func SaveAppConfig(config AppConfig) error {
	return writeConfigToFile(config)
}

func ResetAppConfigToDefault() error {
	_, err := createConfigWithDefaults(configFilePath)
	return err
}

func RevertAppConfigToBackup() error {
	fullConfigPath, err := FullFilePath(configFilePath)
	if err != nil {
		return fmt.Errorf("error getting config path: %w", err)
	}
	fullBackupConfigPath, err := FullFilePath(backupConfigFilePath)
	if err != nil {
		return fmt.Errorf("error getting backup path: %w", err)
	}

	// delete the file if it exists
	if err := os.Remove(fullConfigPath); !os.IsNotExist(err) && err != nil {
		return err
	}
	config, err := loadExistingConfig(fullBackupConfigPath)
	if err != nil {
		return err
	}
	return writeConfigToFile(config)
}

func createConfigWithDefaults(filePath string) (AppConfig, error) {
	config := AppConfig{}
	err := yaml.Unmarshal(embeddedConfigFile, &config)
	if err != nil {
		return config, fmt.Errorf("error unmarshalling embedded config: %w", err)
	}
	// set default model to legacy option (for backwards compat)
	modelOverride := os.Getenv("OPENAI_MODEL_OVERRIDE")
	if modelOverride != "" {
		config.Preferences.DefaultModel = modelOverride
	}

	return config, writeConfigToFile(config)
}

func loadExistingConfig(filePath string) (AppConfig, error) {
	config := AppConfig{}
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return config, fmt.Errorf("error reading config file: %w", err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, fmt.Errorf("error unmarshalling config file: %w", err)
	}
	return config, nil
}

func SaveBackupConfig(config AppConfig) error {
	filePath, err := FullFilePath(backupConfigFilePath)
	if err != nil {
		return fmt.Errorf("error getting backup path: %w", err)
	}
	configData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}
	err = os.WriteFile(filePath, configData, 0644)
	if err != nil {
		return fmt.Errorf("error writing backup config: %w", err)
	}
	return nil
}

func writeConfigToFile(config AppConfig) error {
	filePath, err := FullFilePath(configFilePath)
	if err != nil {
		return fmt.Errorf("error getting config file path: %w", err)
	}
	// Create all directories in the filepath
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directories: %w", err)
	}
	configData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	err = os.WriteFile(filePath, configData, 0644)
	if err != nil {
		return fmt.Errorf("error writing config to file: %w", err)
	}
	// Reuse already-marshaled data for backup to avoid double marshaling
	backupPath, err := FullFilePath(backupConfigFilePath)
	if err == nil {
		_ = os.WriteFile(backupPath, configData, 0644)
	}
	return nil
}
