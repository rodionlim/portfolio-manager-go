package config

import (
	"encoding/json"
	"os"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration.
type Config struct {
	VerboseLogging bool   `yaml:"verboseLogging"`
	LogFilePath    string `yaml:"logFilePath"`
	Host           string `yaml:"host"`
	Port           string `yaml:"port"`
}

// Implement the Stringer interface for Config
func (c Config) String() string {
	jConfig, _ := json.MarshalIndent(c, "", "\t")
	return string(jConfig)
}

// NewConfig reads the configuration file at the given path and returns a Config object.
func NewConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := Config{}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	// Set default value for Host if not provided
	if config.Host == "" {
		config.Host = "localhost"
	}

	return &config, nil
}
