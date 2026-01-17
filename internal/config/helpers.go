package config

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"
)

func getEnvOrPanic(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("[CONFIG] %s not set", key))
	}
	return val
}

func getEnvInt(key string, defaultValue int) int {
	valueString := os.Getenv(key)
	if valueString == "" {
		fmt.Fprintf(os.Stdout, "[CONFIG] %s not set, using default: %d\n", key, defaultValue)
		return defaultValue
	}

	// Atoi(str) is equivalent to ParseInt(str, 10, 0)
	value, err := strconv.Atoi(valueString)
	if err != nil {
		fmt.Fprintf(os.Stdout, "[CONFIG] %s: %s invalid, using default: %v\n", key, valueString, defaultValue)
		return defaultValue
	}

	return value
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	valueString := os.Getenv(key)
	if valueString == "" {
		fmt.Fprintf(os.Stdout, "[CONFIG] %s not set, using default: %s\n", key, defaultValue)
		return defaultValue
	}

	value, err := time.ParseDuration(valueString)
	if err != nil {
		fmt.Fprintf(os.Stdout, "[CONFIG] %s: %s invalid, using default: %v\n", key, valueString, defaultValue)
		return defaultValue
	}

	return value
}

var KEY_LENGTH uint32

func getEnvSecretKey(key string) []byte {
	valueString := os.Getenv(key)
	if valueString == "" {
		panic(fmt.Sprintf("[CONFIG] %s not set", key))
	}

	decoded, err := hex.DecodeString(valueString)
	if err != nil {
		panic(fmt.Sprintf("[CONFIG] decode error: %v", err))
	}

	if len(decoded) != int(KEY_LENGTH) {
		panic(fmt.Sprintf("[CONFIG] %s length is %d bytes; want %d bytes", key, len(decoded), int(KEY_LENGTH)))
	}

	return decoded
}

type uint interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

func getEnvUint[T uint](key string) T {
	valueString, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("[CONFIG] %s not set", key))
	}

	var bitSize int
	switch any(*new(T)).(type) {
	case uint8:
		bitSize = 8
	case uint16:
		bitSize = 16
	case uint32:
		bitSize = 32
	case uint64:
		bitSize = 64
	default:
		// I miss Rust sometimes, she would never do this...
		panic(fmt.Sprintf("[CONFIG] Unsupported type as generic: %v", *new(T)))
	}

	value, err := strconv.ParseUint(valueString, 10, bitSize)
	if err != nil {
		panic(fmt.Sprintf("[CONFIG] Failed to parse: %s", valueString))
	}

	return T(value)
}
