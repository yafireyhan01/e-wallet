package config

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type ApiConfig struct {
	ApiPort string
}

type DbConfig struct {
	Host    string
	Port    string
	Name    string
	User    string
	Pass    string
	Driver  string
	JwtLife int
}

type LogFileConfig struct {
	FilePath string
}

type TokenConfig struct {
	IssuerName      string
	JwtSignatureKey []byte
	JwtLifeTime     time.Duration
}

type Config struct {
	ApiConfig
	DbConfig
	LogFileConfig
	TokenConfig
}

func (c *Config) readConfig() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	lifetime, _ := strconv.Atoi(os.Getenv("TOKEN_LIFE_TIME"))
	c.ApiConfig = ApiConfig{
		ApiPort: os.Getenv("API_PORT"),
	}

	c.DbConfig = DbConfig{
		Host:    os.Getenv("DB_HOST"),
		Port:    os.Getenv("DB_PORT"),
		Name:    os.Getenv("DB_NAME"),
		User:    os.Getenv("DB_USER"),
		Driver:  os.Getenv("DB_DRIVER"),
		Pass:    os.Getenv("DB_PASSWORD"),
		JwtLife: lifetime,
	}

	c.LogFileConfig = LogFileConfig{
		FilePath: os.Getenv("LOG_FILE"),
	}

	c.TokenConfig = TokenConfig{
		IssuerName:      os.Getenv("TOKEN_ISSUE_NAME"),
		JwtSignatureKey: []byte(os.Getenv("TOKEN_KEY")),
		JwtLifeTime:     time.Duration(c.JwtLife) * time.Hour,
	}

	if c.ApiPort == "" || c.Host == "" || c.Port == "" || c.Name == "" || c.User == "" || c.FilePath == "" || c.IssuerName == "" ||
		c.JwtSignatureKey == nil || c.JwtLifeTime == 0 {
		return errors.New("environment required")
	}

	return nil
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := cfg.readConfig(); err != nil {
		return nil, err
	}

	return cfg, nil
}
