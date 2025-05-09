package config

import (
	"encoding/json"
	"errors"
	"os"
	root "portfolio-manager"
	"sync"

	"portfolio-manager/internal/dal"
	"portfolio-manager/pkg/common"

	"gopkg.in/yaml.v2"
)

// DividendsConfig nests all dividend-related configuration
// Withholding tax rates are in decimal (e.g., 0.15 for 15%)
type DividendsConfig struct {
	WithholdingTaxSG float64 `yaml:"withholdingTaxSG"`
	WithholdingTaxUS float64 `yaml:"withholdingTaxUS"`
	WithholdingTaxHK float64 `yaml:"withholdingTaxHK"`
	WithholdingTaxIE float64 `yaml:"withholdingTaxIE"`
}

// MetricsConfig nests all metrics-related configuration
// Schedule is a cron string for metrics collection
type MetricsConfig struct {
	Schedule string `yaml:"schedule"`
}

// Config represents the application configuration.
type Config struct {
	VerboseLogging  bool            `yaml:"verboseLogging"`
	LogFilePath     string          `yaml:"logFilePath"`
	Host            string          `yaml:"host"`
	Port            string          `yaml:"port"`
	BaseCcy         string          `yaml:"baseCcy"`
	Db              string          `yaml:"db"`
	DbPath          string          `yaml:"dbPath"`
	RefDataSeedPath string          `yaml:"refDataSeedPath"`
	Dividends       DividendsConfig `yaml:"dividends"`
	Metrics         MetricsConfig   `yaml:"metrics"`
}

// Implement the Stringer interface for Config.
func (c Config) String() string {
	jConfig, _ := json.MarshalIndent(c, "", "\t")
	return string(jConfig)
}

// Singleton instance variables.
var (
	instance *Config
	once     sync.Once
	initErr  error
)

// SetConfig sets the singleton Config instance (useful for testing).
func SetConfig(cfg *Config) {
	instance = cfg
}

// initializeConfig handles unmarshalling, setting defaults and validations.
// It assigns to the package-level 'instance'.
func initializeConfig(data []byte) error {
	cfg := Config{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	// Set default for Host if not provided.
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}

	// Set default & validate the database field.
	if cfg.Db == "" {
		cfg.Db = dal.LDB
	}
	if cfg.Db != dal.LDB && cfg.Db != dal.RDB {
		return errors.New("invalid db type: must be 'leveldb' or 'rocksdb'")
	}
	// Set default for DbPath if not provided.
	if cfg.DbPath == "" {
		cfg.DbPath = "./portfolio-manager.db"
	}

	// Set defaults for DividendsConfig if not provided
	// (all default to 0)

	// Set default for MetricsConfig if not provided
	if cfg.Metrics.Schedule == "" {
		cfg.Metrics.Schedule = "10 17 * * 1-5" // default: 5:10pm Mon-Fri
	}

	instance = &cfg
	return nil
}

// GetOrCreateConfig reads configuration from the provided file path, provided it has not been set before
func GetOrCreateConfig(path string) (*Config, error) {
	once.Do(func() {
		if instance != nil {
			return
		}
		var data []byte
		data, initErr = os.ReadFile(path)
		if initErr != nil {
			// Attempt to read from embeddedFS if file not found.
			data, initErr = root.EmbeddedFiles.ReadFile(common.SanitizePath(path))
			if initErr != nil {
				return
			}
		}
		initErr = initializeConfig(data)
	})
	return instance, initErr
}
