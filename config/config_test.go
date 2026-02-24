package config

import (
	"os"
	"path/filepath"
	"q/types"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestEmbeddedConfigParsesCorrectly(t *testing.T) {
	var config AppConfig
	err := yaml.Unmarshal(embeddedConfigFile, &config)
	if err != nil {
		t.Fatalf("failed to unmarshal embedded config: %v", err)
	}
	if len(config.Models) == 0 {
		t.Error("embedded config should have at least one model")
	}
	if config.Preferences.DefaultModel == "" {
		t.Error("embedded config should have a default model")
	}
	if config.Version == "" {
		t.Error("embedded config should have a version")
	}
}

func TestLoadAndSaveConfig(t *testing.T) {
	// Use temp dir to avoid touching real config
	tmpDir := t.TempDir()
	originalConfigPath := configFilePath
	originalBackupPath := backupConfigFilePath

	configFilePath = filepath.Join(tmpDir, "config.yaml")
	backupConfigFilePath = filepath.Join(tmpDir, "backup.yaml")

	// Override FullFilePath to return paths directly
	defer func() {
		configFilePath = originalConfigPath
		backupConfigFilePath = originalBackupPath
	}()

	// Write a test config
	testConfig := AppConfig{
		Preferences: types.Preferences{DefaultModel: "test-model"},
		Version:     "1",
	}

	data, err := yaml.Marshal(testConfig)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(configFilePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configFilePath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Load it back
	loaded, err := loadExistingConfig(configFilePath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if loaded.Preferences.DefaultModel != "test-model" {
		t.Errorf("default model = %q, want %q", loaded.Preferences.DefaultModel, "test-model")
	}
}

func TestFullFilePath(t *testing.T) {
	path, err := FullFilePath("test/file.yaml")
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, "test/file.yaml")
	if path != expected {
		t.Errorf("got %q, want %q", path, expected)
	}
}
