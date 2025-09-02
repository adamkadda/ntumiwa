package config

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

/*
	Thank you u/jxstack on Reddit for their advice on
	good environment variable practices.

	https://www.reddit.com/r/golang/comments/1dzxah6/comment/lcjfw2h
*/

type Config struct {
	API        APIClientConfig
	DB         PostgresConfig
	Session    SessionManagerConfig
	Logging    LogConfig
	Cache      CacheConfig
	Hash       HashConfig
	Host       string
	Port       string
	ServerType string
	SecretKey  []byte
}

type LogConfig struct {
	Style string
	Level string
}

type CacheConfig struct {
	TimeToLive time.Duration
	GCInterval time.Duration
}

type PostgresConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	DBName   string
}

type APIClientConfig struct {
	BaseURL string
	Timeout time.Duration

	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
}

type SessionManagerConfig struct {
	GCInterval         time.Duration
	IdleExpiration     time.Duration
	AbsoluteExpiration time.Duration
	CookieName         string
	Domain             string
}

type HashConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func (pg *PostgresConfig) Connect() (*sql.DB, error) {
	if pg.DBName == "" {
		return nil, nil
	}

	connStr := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		pg.Username,
		pg.Password,
		pg.Host,
		pg.Port,
		pg.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}

/*
	This method returns an error for future-proofing.
	In the event that some part of the startup fails e.g. missing DB
	credentials, or log specs, we can return an appropriate error.

	Still deciding if this is worth doing ^^, error types and all
*/

func LoadConfig() (*Config, error) {
	config := &Config{
		Host:       os.Getenv("HOST"),
		Port:       os.Getenv("PORT"),
		ServerType: os.Getenv("SERVER_TYPE"),
		DB: PostgresConfig{
			DBName:   os.Getenv("POSTGRES_DB"),
			Username: os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PWD"),
			Host:     os.Getenv("POSTGRES_HOST"),
			Port:     os.Getenv("POSTGRES_PORT"),
		},
		API: APIClientConfig{
			BaseURL:             os.Getenv("API_BASE_URL"),
			Timeout:             getEnvDuration("API_TIMEOUT", 500*time.Millisecond),
			IdleConnTimeout:     getEnvDuration("API_IDLE_CONN_TIMEOUT", 60*time.Second),
			MaxIdleConns:        getEnvInt("API_MAX_IDLE_CONNS", 100),
			MaxIdleConnsPerHost: getEnvInt("API_MAX_IDLE_CONNS_PER_HOST", 100),
		},
		Cache: CacheConfig{},
		Session: SessionManagerConfig{
			GCInterval:         getEnvDuration("GC_INTERVAL", 1*time.Hour),
			IdleExpiration:     getEnvDuration("SESSION_TTI", 1*time.Hour),
			AbsoluteExpiration: getEnvDuration("SESSION_TTL", 8*time.Hour),
			CookieName:         os.Getenv("SESSION_COOKIE_NAME"),
			Domain:             os.Getenv("DOMAIN"),
		},
		Hash: HashConfig{
			Memory:      getEnvUint[uint32]("MEMORY"),
			Iterations:  getEnvUint[uint32]("ITERATIONS"),
			Parallelism: getEnvUint[uint8]("PARALLELISM"),
			SaltLength:  getEnvUint[uint32]("SALT_LENGTH"),
			KeyLength:   getEnvUint[uint32]("KEY_LENGTH"),
		},
	}

	KEY_LENGTH = config.Hash.KeyLength
	config.SecretKey = getEnvSecretKey("SECRET_KEY")

	if config.ServerType == "" {
		panic("[CONFIG] SERVER_TYPE not set")
	}

	if config.ServerType == "API" {
		if config.SessionMan.CookieName == "" {
			panic("[CONFIG] COOKIE_NAME not set")
		}

		if config.SessionMan.Domain == "" {
			panic("[CONFIG] DOMAIN not set")
		}

	} else if config.API.BaseURL == "" {
		panic("[ERROR] API_BASE_URL not set")
	}

	return config, nil
}
