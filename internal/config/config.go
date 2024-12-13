package config

import (
	"encoding/json"
	"errors"
	"os"
	"portfolio-manager/internal/dal"
	"sync"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration.
type Config struct {
	VerboseLogging     bool    `yaml:"verboseLogging"`
	LogFilePath        string  `yaml:"logFilePath"`
	Host               string  `yaml:"host"`
	Port               string  `yaml:"port"`
	Db                 string  `yaml:"db"`
	DbPath             string  `yaml:"dbPath"`
	RefDataSeedPath    string  `yaml:"refDataSeedPath"`
	DivWitholdingTaxSG float64 `yaml:"divWitholdingTaxSG"`
	DivWitholdingTaxUS float64 `yaml:"divWitholdingTaxUS"`
	DivWitholdingTaxHK float64 `yaml:"divWitholdingTaxHK"`
}

// Implement the Stringer interface for Config
func (c Config) String() string {
	jConfig, _ := json.MarshalIndent(c, "", "\t")
	return string(jConfig)
}

var (
	instance *Config
	once     sync.Once
	err      error
)

// GetOrCreateConfig returns the singleton Config instance, and instantiates it if it hasn't already been done so.
func GetOrCreateConfig(path string) (*Config, error) {
	once.Do(func() {
		var file []byte
		file, err = os.ReadFile(path)
		if err != nil {
			return
		}

		config := Config{}
		err = yaml.Unmarshal(file, &config)
		if err != nil {
			return
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
			err = errors.New("invalid db type: must be 'leveldb' or 'rocksdb'")
			return
		}
		if config.DbPath == "" {
			config.DbPath = "./portfolio-manager.db"
		}

		instance = &config
	})

	return instance, err
}
