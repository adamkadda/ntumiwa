package config

import (
	"fmt"
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
	API       APIClientConfig
	DB        PostgresConfig
	Session   SessionManagerConfig
	Logging   LogConfig
	Hash      HashConfig
	Host      string
	Port      string
	Kind      string
	SecretKey []byte
}

type LogConfig struct {
	Style string
}

type PostgresConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	DBName   string
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
		c.DBName,
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

func LoadConfig() (*Config, error) {
	config := &Config{
		Host: os.Getenv("HOST"),
		Port: os.Getenv("PORT"),
		Kind: os.Getenv("KIND"),
		DB: PostgresConfig{
			DBName:                    os.Getenv("POSTGRES_DB"),
			Username:                  os.Getenv("POSTGRES_USER"),
			Password:                  os.Getenv("POSTGRES_PWD"),
			Host:                      os.Getenv("POSTGRES_HOST"),
			Port:                      os.Getenv("POSTGRES_PORT"),
			SSLMode:                   os.Getenv("POSTGRES_SSLMODE"),
			Timeout:                   getEnvDuration("POSTGRES_TIMEOUT", 1*time.Minute),
			PoolMaxConns:              getEnvInt("POSTGRES_POOL_MAX_CONNS", 5),
			PoolMinConns:              getEnvInt("POSTGRES_POOL_MIN_CONNS", 1),
			PoolMaxConnLifetime:       os.Getenv("POSTGRES_POOL_MAX_CONN_LIFETIME"),
			PoolMaxConnIdleTime:       os.Getenv("POSTGRES_POOL_MAX_CONN_IDLE_TIME"),
			PoolHealthCheckPeriod:     os.Getenv("POSTGRES_POOL_HEALTH_CHECK_PERIOD"),
			PoolMaxConnLifetimeJitter: os.Getenv("POSTGRES_POOL_MAX_CONN_LIFETIME_JITTER"),
		},
		API: APIClientConfig{
			BaseURL:             os.Getenv("API_BASE_URL"),
			Timeout:             getEnvDuration("API_TIMEOUT", 500*time.Millisecond),
			IdleConnTimeout:     getEnvDuration("API_IDLE_CONN_TIMEOUT", 60*time.Second),
			MaxIdleConns:        getEnvInt("API_MAX_IDLE_CONNS", 100),
			MaxIdleConnsPerHost: getEnvInt("API_MAX_IDLE_CONNS_PER_HOST", 100),
		},
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
		Logging: LogConfig{
			Style: os.Getenv("LOG_STYLE"),
		},
	}

	config.SecretKey = getEnvSecretKey("SECRET_KEY")

	// NOTE: Validation
	if config.Kind == "" {
		panic("[CONFIG] KIND not set")
	}

	if config.Kind == "API" {
		if config.Session.CookieName == "" {
			panic("[CONFIG] SESSION_COOKIE_NAME not set")
		}
		if config.Session.Domain == "" {
			panic("[CONFIG] DOMAIN not set")
		}
	} else if config.API.BaseURL == "" {
		panic("[CONFIG] API_BASE_URL not set")
	}

	return config, nil
}
