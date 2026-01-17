package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_AllFieldsValid(t *testing.T) {
	os.Setenv("HOST", "localhost")
	os.Setenv("PORT", "8080")
	os.Setenv("SERVER_TYPE", "PUBLIC")

	os.Setenv("POSTGRES_DB", "mydb")
	os.Setenv("POSTGRES_USER", "user")
	os.Setenv("POSTGRES_PWD", "pwd")
	os.Setenv("POSTGRES_HOST", "dbhost")
	os.Setenv("POSTGRES_PORT", "5432")

	os.Setenv("API_BASE_URL", "http://example.com")
	os.Setenv("API_TIMEOUT", "600ms")
	os.Setenv("API_IDLE_CONN_TIMEOUT", "90s")
	os.Setenv("API_MAX_IDLE_CONNS", "10")
	os.Setenv("API_MAX_IDLE_CONNS_PER_HOST", "5")

	defer os.Clearenv()

	config, err := LoadConfig()

	assert.Nil(t, err)
	assert.NotNil(t, config)

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "8080", config.Port)
	assert.Equal(t, "PUBLIC", config.ServerType)

	assert.Equal(t, "mydb", config.DB.DBName)
	assert.Equal(t, "user", config.DB.Username)
	assert.Equal(t, "pwd", config.DB.Password)
	assert.Equal(t, "dbhost", config.DB.Host)
	assert.Equal(t, "5432", config.DB.Port)

	assert.NotEqual(t, "", config.API.BaseURL)

	expectedTimeout, _ := time.ParseDuration("600ms")
	assert.Equal(t, expectedTimeout, config.API.Timeout)

	expectedIdleTimeout, _ := time.ParseDuration("90s")
	assert.Equal(t, expectedIdleTimeout, config.API.IdleConnTimeout)

	assert.Equal(t, 10, config.API.MaxIdleConns)
	assert.Equal(t, 5, config.API.MaxIdleConnsPerHost)
}

func TestLoadConfig_ServerTypeNotSet(t *testing.T) {
	os.Setenv("HOST", "localhost")
	os.Setenv("PORT", "8080")
	// os.Setenv("SERVER_TYPE", "")

	os.Setenv("POSTGRES_DB", "mydb")
	os.Setenv("POSTGRES_USER", "user")
	os.Setenv("POSTGRES_PWD", "pwd")
	os.Setenv("POSTGRES_HOST", "dbhost")
	os.Setenv("POSTGRES_PORT", "5432")

	os.Setenv("API_BASE_URL", "http://example.com")
	os.Setenv("API_TIMEOUT", "600ms")
	os.Setenv("API_IDLE_CONN_TIMEOUT", "90s")
	os.Setenv("API_MAX_IDLE_CONNS", "10")
	os.Setenv("API_MAX_IDLE_CONNS_PER_HOST", "5")

	defer os.Clearenv()

	config, err := LoadConfig()

	assert.Nil(t, config)
	assert.EqualError(t, err, "[CONFIG] SERVER_TYPE not set")
}

func TestLoadConfig_APIBaseUrlNotSet(t *testing.T) {
	os.Setenv("HOST", "localhost")
	os.Setenv("PORT", "8080")
	os.Setenv("SERVER_TYPE", "PUBLIC")

	os.Setenv("POSTGRES_DB", "mydb")
	os.Setenv("POSTGRES_USER", "user")
	os.Setenv("POSTGRES_PWD", "pwd")
	os.Setenv("POSTGRES_HOST", "dbhost")
	os.Setenv("POSTGRES_PORT", "5432")

	// os.Setenv("API_BASE_URL", "http://example.com")
	os.Setenv("API_TIMEOUT", "600ms")
	os.Setenv("API_IDLE_CONN_TIMEOUT", "90s")
	os.Setenv("API_MAX_IDLE_CONNS", "10")
	os.Setenv("API_MAX_IDLE_CONNS_PER_HOST", "5")

	defer os.Clearenv()

	config, err := LoadConfig()

	assert.Nil(t, config)
	assert.EqualError(t, err, "[API_CONFIG] API_BASE_URL not set")
}
