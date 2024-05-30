package config

import (
	"os"
	"strconv"

	"github.com/ava-labs/hypersdk/codec"
	"github.com/nuklai/nuklaivm/consts"
)

type Config struct {
	HTTPHost string
	HTTPPort int

	NuklaiRPC string

	Recipient     string
	recipientAddr codec.Address

	FeedSize               int
	MinFee                 uint64
	FeeDelta               uint64
	MessagesPerEpoch       int
	TargetDurationPerEpoch int64 // seconds

	AdminToken string

	// PostgreSQL configuration
	PostgresHost     string
	PostgresPort     int
	PostgresUser     string
	PostgresPassword string
	PostgresDBName   string
	PostgresSSLMode  string
}

func (c *Config) RecipientAddress() (codec.Address, error) {
	if c.recipientAddr != codec.EmptyAddress {
		return c.recipientAddr, nil
	}
	addr, err := codec.ParseAddressBech32(consts.HRP, c.Recipient)
	if err == nil {
		c.recipientAddr = addr
	}
	return addr, err
}

func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func LoadConfigFromEnv() (*Config, error) {
	port, err := strconv.Atoi(GetEnv("PORT", "10592"))
	if err != nil {
		return nil, err
	}

	feedSize, err := strconv.Atoi(GetEnv("FEEDSIZE", "100"))
	if err != nil {
		return nil, err
	}

	minFee, err := strconv.ParseUint(GetEnv("MIN_FEE", "1000000"), 10, 64)
	if err != nil {
		return nil, err
	}

	feeDelta, err := strconv.ParseUint(GetEnv("FEE_DELTA", "100000"), 10, 64)
	if err != nil {
		return nil, err
	}

	messagesPerEpoch, err := strconv.Atoi(GetEnv("MESSAGES_PER_EPOCH", "100"))
	if err != nil {
		return nil, err
	}

	targetDurationPerEpoch, err := strconv.ParseInt(GetEnv("TARGET_DURATION_PER_EPOCH", "300"), 10, 64)
	if err != nil {
		return nil, err
	}

	postgresPort, err := strconv.Atoi(GetEnv("POSTGRES_PORT", "5432"))
	if err != nil {
		return nil, err
	}

	postgresEnableSSL := GetEnv("POSTGRES_ENABLESSL", "false")
	postgresSSLMode := "disable"
	if parsed, err := strconv.ParseBool(postgresEnableSSL); err == nil && parsed {
		postgresSSLMode = "require"
	}

	return &Config{
		HTTPHost: GetEnv("HOST", ""),
		HTTPPort: port,

		NuklaiRPC: os.Getenv("NUKLAI_RPC"),

		Recipient:              GetEnv("RECIPIENT", "nuklai1qpg4ecapjymddcde8sfq06dshzpxltqnl47tvfz0hnkesjz7t0p35d5fnr3"),
		FeedSize:               feedSize,
		MinFee:                 minFee,
		FeeDelta:               feeDelta,
		MessagesPerEpoch:       messagesPerEpoch,
		TargetDurationPerEpoch: targetDurationPerEpoch,

		AdminToken: GetEnv("ADMIN_TOKEN", "ADMIN_TOKEN"),

		PostgresHost:     GetEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     postgresPort,
		PostgresUser:     GetEnv("POSTGRES_USER", "user"),
		PostgresPassword: GetEnv("POSTGRES_PASSWORD", "password"),
		PostgresDBName:   GetEnv("POSTGRES_DBNAME", "dbname"),
		PostgresSSLMode:  postgresSSLMode,
	}, nil
}
