package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

/*
	Thank you u/jxstack on Reddit for their advice on
	good environment variable practices.

	https://www.reddit.com/r/golang/comments/1dzxah6/comment/lcjfw2h
*/

type Config struct {
	Host    string
	Port    string
	AppEnv  string
	DB      PostgresConfig
	Hash    HashConfig
	Client  ClientConfig
	Session SessionConfig
	Logging LogConfig
}

type LogConfig struct {
	Style string // e.g. json, text
	Level string // e.g. debug, info, warn, error
}

type PostgresConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	Name     string
	SSLMode  string
	Timeout  time.Duration

	// Pool settings
	PoolMaxConns              int    // > 0
	PoolMinConns              int    // >= 0
	PoolMaxConnLifetime       string // duration string, e.g. "1h", "1h30m"
	PoolMaxConnIdleTime       string // duration string, e.g. "30m"
	PoolHealthCheckPeriod     string // duration string, e.g. "1m"
	PoolMaxConnLifetimeJitter string // duration string, e.g. "10s"
}

func (c PostgresConfig) DSN() string {
	// Base: postgres://user:pass@host:port/dbname
	password := url.QueryEscape(c.Password)

	base := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		c.Username,
		password,
		c.Host,
		c.Port,
		c.Name,
	)

	params := url.Values{}

	if c.SSLMode != "" {
		params.Add("sslmode", c.SSLMode)
	}
	if c.PoolMaxConns > 0 {
		params.Add("pool_max_conns", fmt.Sprintf("%d", c.PoolMaxConns))
	}
	if c.PoolMinConns >= 0 {
		params.Add("pool_min_conns", fmt.Sprintf("%d", c.PoolMinConns))
	}
	if c.PoolMaxConnLifetime != "" {
		params.Add("pool_max_conn_lifetime", c.PoolMaxConnLifetime)
	}
	if c.PoolMaxConnIdleTime != "" {
		params.Add("pool_max_conn_idle_time", c.PoolMaxConnIdleTime)
	}
	if c.PoolHealthCheckPeriod != "" {
		params.Add("pool_health_check_period", c.PoolHealthCheckPeriod)
	}
	if c.PoolMaxConnLifetimeJitter != "" {
		params.Add("pool_max_conn_lifetime_jitter", c.PoolMaxConnLifetimeJitter)
	}

	// Append params if any exist
	if len(params) > 0 {
		base = base + "?" + strings.ReplaceAll(params.Encode(), "+", "")
	}

	return base
}

type ClientConfig struct {
	BaseURL             string
	Timeout             time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
}

type SessionConfig struct {
	GCInterval         time.Duration
	IdleExpiration     time.Duration
	AbsoluteExpiration time.Duration

	// NOTE: Cookie configuration
	CookieName string
	Domain     string
	SecretKey  []byte // for signing cookies
}

type HashConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func LoadConfig() (*Config, error) {
	host := getEnvOrPanic("HOST")
	port := getEnvOrPanic("PORT")

	appEnv := strings.ToUpper(getEnvOrPanic("APP_ENV"))
	switch appEnv {
	case "DEV":
	case "TEST":
	case "STAGING":
	case "PROD":
	default:
		log.Fatalf("[CONFIG] APP_ENV=%s invalid", appEnv)
	}

	logging := LogConfig{
		Style: getEnvOrPanic("LOG_STYLE"),
		Level: getEnvOrPanic("LOG_LEVEL"),
	}

	var db PostgresConfig
	var session SessionConfig
	var hash HashConfig
	var client ClientConfig

	kind := getEnvOrPanic("KIND")
	switch kind {
	case "API":
		db = PostgresConfig{
			Name:     getEnvOrPanic("POSTGRES_DB"),
			Username: getEnvOrPanic("POSTGRES_USER"),
			Password: getEnvOrPanic("POSTGRES_PWD"),
			Host:     getEnvOrPanic("POSTGRES_HOST"),
			Port:     getEnvOrPanic("POSTGRES_PORT"),
			SSLMode:  getEnvOrPanic("POSTGRES_SSLMODE"),
			Timeout:  getEnvDuration("POSTGRES_TIMEOUT", 1*time.Minute),

			// NOTE: Optional but recommended pool configuration
			PoolMaxConns:              getEnvInt("POSTGRES_POOL_MAX_CONNS", 5),
			PoolMinConns:              getEnvInt("POSTGRES_POOL_MIN_CONNS", 1),
			PoolMaxConnLifetime:       os.Getenv("POSTGRES_POOL_MAX_CONN_LIFETIME"),
			PoolMaxConnIdleTime:       os.Getenv("POSTGRES_POOL_MAX_CONN_IDLE_TIME"),
			PoolHealthCheckPeriod:     os.Getenv("POSTGRES_POOL_HEALTH_CHECK_PERIOD"),
			PoolMaxConnLifetimeJitter: os.Getenv("POSTGRES_POOL_MAX_CONN_LIFETIME_JITTER"),
		}

		KEY_LENGTH = getEnvUint[uint32]("KEY_LENGTH")

		hash = HashConfig{
			Memory:      getEnvUint[uint32]("MEMORY") * 1024, // MiB -> KiB
			Iterations:  getEnvUint[uint32]("ITERATIONS"),
			Parallelism: getEnvUint[uint8]("PARALLELISM"),
			SaltLength:  getEnvUint[uint32]("SALT_LENGTH"),
			KeyLength:   KEY_LENGTH,
		}

		if appEnv != "TEST" {
			session = SessionConfig{
				GCInterval:         getEnvDuration("GC_INTERVAL", 1*time.Hour),
				IdleExpiration:     getEnvDuration("SESSION_TTI", 1*time.Hour),
				AbsoluteExpiration: getEnvDuration("SESSION_TTL", 8*time.Hour),
				CookieName:         getEnvOrPanic("SESSION_COOKIE_NAME"),
				Domain:             getEnvOrPanic("SESSION_COOKIE_DOMAIN"),
				SecretKey:          getEnvSecretKey("SECRET_KEY"),
			}
		}
	case "ADMIN", "PUBLIC":
		client = ClientConfig{
			BaseURL:             getEnvOrPanic("API_BASE_URL"),
			Timeout:             getEnvDuration("API_TIMEOUT", 500*time.Millisecond),
			IdleConnTimeout:     getEnvDuration("API_IDLE_CONN_TIMEOUT", 60*time.Second),
			MaxIdleConns:        getEnvInt("API_MAX_IDLE_CONNS", 100),
			MaxIdleConnsPerHost: getEnvInt("API_MAX_IDLE_CONNS_PER_HOST", 100),
		}
	default:
		log.Fatalf("[CONFIG] KIND=%s invalid", kind)
	}

	config := &Config{
		Host:    host,
		Port:    port,
		AppEnv:  appEnv,
		DB:      db,
		Hash:    hash,
		Client:  client,
		Session: session,
		Logging: logging,
	}

	return config, nil
}
