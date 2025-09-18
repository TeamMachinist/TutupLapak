package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Redis    RedisConfig
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type AppConfig struct {
	Port       string
	APITimeout time.Duration
	FileUrl    string
	Env        string
}

type DatabaseConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	DBName      string
	SSLMode     string
	DatabaseURL string
}

type JWTConfig struct {
	Secret   string        `json:"secret"`
	Duration time.Duration `json:"duration"`
	Issuer   string        `json:"issuer"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	jwtDurationStr := getEnv("JWT_DURATION", "24h")
	jwtDuration, _ := time.ParseDuration(jwtDurationStr)
	if jwtDuration == 0 {
		jwtDuration = 24 * time.Hour
	}

	rdb, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		fmt.Print("REDIS DB NOT FOUND")
	}

	config := &Config{
		App: AppConfig{
			Port:    getEnv("PORT", "8002"),
			FileUrl: getEnv("FILES_SERVICE_URL", "http://localhost:8003"),
			Env:     getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST", "localhost"),
			Port:        getEnv("DB_PORT", "5432"),
			User:        getEnv("DB_USER", "postgres"),
			Password:    getEnv("DB_PASSWORD", ""),
			DBName:      getEnv("DB_NAME", "myapp"),
			SSLMode:     getEnv("DB_SSL_MODE", "disable"),
			DatabaseURL: getEnv("DATABASE_URL", ""),
		},
		JWT: JWTConfig{
			Secret:   getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
			Duration: jwtDuration,
			Issuer:   getEnv("JWT_ISSUER", "tutuplapak-core-service"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", "redispass"),
			DB:       rdb,
		},
	}

	timeoutStr := getEnv("API_TIMEOUT", "30s")
	if timeout, err := time.ParseDuration(timeoutStr); err == nil {
		config.App.APITimeout = timeout
	} else {
		config.App.APITimeout = 30 * time.Second
	}

	return config, nil
}

func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}