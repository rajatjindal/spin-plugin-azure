package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
)

type Config struct {
	SubscriptionID    string `json:"subscriptionId"`
	TenantID          string `json:"tenantId"`
	ResourceGroup     string `json:"resourceGroup"`
	ClusterName       string `json:"clusterName"`
	Location          string `json:"location"`
	WorkloadIdentity  string `json:"workloadIdentity"`
	WorkspaceLocation string `json:"workspaceLocation"`
}

func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".spin-azure")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}

// LoadConfig loads the configuration from disk
func LoadConfig() (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := &Config{}
		return defaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(config *Config) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetAzureCredential returns an Azure credential for authentication
func GetAzureCredential() (*azidentity.DefaultAzureCredential, error) {
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}
	return credential, nil
}
