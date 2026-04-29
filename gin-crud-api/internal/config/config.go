package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "release")

	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	if err := viper.ReadInConfig(); err == nil {
		setIfPresent := func(src, dest string) {
			if viper.IsSet(src) {
				viper.Set(dest, viper.Get(src))
			}
		}

		setIfPresent("SERVER_HOST", "server.host")
		setIfPresent("SERVER_PORT", "server.port")
		setIfPresent("SERVER_MODE", "server.mode")

		setIfPresent("DB_HOST", "database.host")
		setIfPresent("DB_PORT", "database.port")
		setIfPresent("DB_USER", "database.user")
		setIfPresent("DB_PASSWORD", "database.password")
		setIfPresent("DB_NAME", "database.dbname")
		setIfPresent("DB_SSLMODE", "database.sslmode")
		setIfPresent("DB_MAX_OPEN_CONNS", "database.max_open_conns")
		setIfPresent("DB_MAX_IDLE_CONNS", "database.max_idle_conns")
		setIfPresent("DB_CONN_MAX_LIFETIME", "database.conn_max_lifetime")

		setIfPresent("LOG_LEVEL", "log.level")
		setIfPresent("LOG_FORMAT", "log.format")
	}

	_ = viper.BindEnv("server.host", "SERVER_HOST")
	_ = viper.BindEnv("server.port", "SERVER_PORT")
	_ = viper.BindEnv("server.mode", "SERVER_MODE")

	_ = viper.BindEnv("database.host", "DB_HOST")
	_ = viper.BindEnv("database.port", "DB_PORT")
	_ = viper.BindEnv("database.user", "DB_USER")
	_ = viper.BindEnv("database.password", "DB_PASSWORD")
	_ = viper.BindEnv("database.dbname", "DB_NAME")
	_ = viper.BindEnv("database.sslmode", "DB_SSLMODE")
	_ = viper.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	_ = viper.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	_ = viper.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")

	_ = viper.BindEnv("log.level", "LOG_LEVEL")
	_ = viper.BindEnv("log.format", "LOG_FORMAT")

	viper.AutomaticEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
