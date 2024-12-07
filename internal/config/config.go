package config

import (
	"encoding/json"
	"errors"
	"os"
	"portfolio-manager/internal/dal"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration.
type Config struct {
	VerboseLogging bool   `yaml:"verboseLogging"`
	LogFilePath    string `yaml:"logFilePath"`
	Host           string `yaml:"host"`
	Port           string `yaml:"port"`
	Db             string `yaml:"db"`
	DbPath         string `yaml:"dbPath"`
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

	// Validate the database field
	if config.Db == "" {
		config.Db = dal.LDB
	}
	if config.Db != dal.LDB && config.Db != dal.RDB {
		return nil, errors.New("invalid db type: must be 'leveldb' or 'rocksdb'")
	}
	if config.DbPath == "" {
		config.DbPath = "./portfolio-manager.db"
	}

	return &config, nil
}
