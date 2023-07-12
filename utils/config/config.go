package config

import "os"

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func LoadRedisConfig() RedisConfig {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost" // Set default host
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379" // Set default port
	}

	password := os.Getenv("REDIS_PASSWORD")
	db := 0 // Set default DB

	return RedisConfig{
		Host:     host,
		Port:     port,
		Password: password,
		DB:       db,
	}
}
