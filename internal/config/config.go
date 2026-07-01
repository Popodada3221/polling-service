package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int
	RedisTTL      int

	ServicePort string
}

func LoadConfig() (*Config, error) {
	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))

	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %v", err)
	}

	redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT: %w", err)
	}

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	redisTTL, err := strconv.Atoi(getEnv("REDIS_TTL", "300"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_TTL: %w", err)
	}

	return &Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        port,
		DBUser:        getEnv("DB_USER", "poll_user"),
		DBPassword:    getEnv("DB_PASSWORD", "poll_password"),
		DBName:        getEnv("DB_NAME", "polling_service"),
		DBSSLMode:     getEnv("DB_SSL_MODE", "disable"),
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     redisPort,
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
		RedisTTL:      redisTTL,
		ServicePort:   getEnv("SERVICE_PORT", "8080"),
	}, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.RedisHost, c.RedisPort)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}
