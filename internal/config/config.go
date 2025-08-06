package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type DB struct {
	Addr     string `mapstructure:"addr"`
	Keyspace string `mapstructure:"keyspace"`
}
type Upload struct {
	Count int
}

type Metrics struct {
	Port string `mapstructure:"port"`
}

type Fetch struct {
	HighPrioretyCount int `mapstructure:"high_priority_count"`
	MedPrioretyCount  int `mapstructure :"med_priority_count"`
	LowPrioretyCount  int `mapstructure:"low_priority_count"`
}

type Scripts struct {
	Access string `mapstructure:"access"`
	Update string `mapstructure:"update"`
}

type Queue struct {
	Stream     string `mapstructure:"stream"`
	GroupName  string `mapstructure:"group_name"`
	ConsumerID string `mapstructure:"consumer_id"`
}

type Workers struct {
	Upload Upload `mapstructure:"upload"`
	Fetch  Fetch  `mapstructure:"fetch"`
	Total  int    `mapstructure:"total"`
}

type Cache struct {
	Addr string `mapstructure:"addr"`
}

type Config struct {
	Metrics        *Metrics `mapstructure:"metrics"`
	Scripts        *Scripts `mapstructure:"scripts_path"`
	Workers        Workers  `mapstructure:"workers"`
	Queue          *Queue   `mapstructure:"queue"`
	Cache          *Cache   `mapstructure:"cache"`
	MaxConcurrency int      `mapstructure:"max_concurrency"`
	DB             *DB      `mapstructure:"db"`
}

func NewConfig() (*Config, error) {
	setDefault()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("fatal error reading config file: %w", err)
		}
		fmt.Println("Config file not found, using default values.")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefault() {
	viper.SetDefault("max_concurrency", 5)

	viper.SetDefault("queue.stream", "tasks")
	viper.SetDefault("queue.group_name", "default_group")     // Added default for group_name
	viper.SetDefault("queue.consumer_id", "default_consumer") // Added default for consumer_id

	viper.SetDefault("cache.addr", "localhost:9042")

	viper.SetDefault("db.addr", "localhost:9093")
	viper.SetDefault("db.keyspace", "default_keyspace") // Added default for keyspace

	viper.SetDefault("scripts_path.access", "/default/access_script.sh")
	viper.SetDefault("scripts_path.update", "/default/update_script.sh")

	viper.SetDefault("metrics.port", ":2112")

	viper.SetDefault("workers.total", 50)
	viper.SetDefault("workers.upload.count", 5)
	viper.SetDefault("workers.fetch.high_priority_count", 20)
	viper.SetDefault("workers.fetch.med_priority_count", 15)
	viper.SetDefault("workers.fetch.low_priority_count", 10)
}
