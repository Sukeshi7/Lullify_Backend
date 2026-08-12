package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port             string
	Env              string
	DatabaseURL      string
	RedisURL         string
	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
	OTELEndpoint     string
	OTELServiceName  string
	StorageProvider  string
	StoragePath      string

	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool

	MaxUploadSizeBytes int64
}

func Load() *Config {
	return &Config{
		Port:             getEnv("PORT", "8080"),
		Env:              getEnv("ENV", "development"),
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://lullify:password@localhost:5432/lullify?sslmode=disable"),
		RedisURL:         getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "changeme-access"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "changeme-refresh"),
		JWTAccessExpiry:  parseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m")),
		JWTRefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h")),
		OTELEndpoint:     getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
		OTELServiceName:  getEnv("OTEL_SERVICE_NAME", "lullify-backend"),
		StorageProvider:  getEnv("STORAGE_PROVIDER", "minio"),
		StoragePath:      getEnv("STORAGE_PATH", "/data/audio"),

		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", "lullify_admin"),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", "lullify_dev_secret_change_me"),
		MinIOBucket:    getEnv("MINIO_BUCKET", "lullify-audio"),
		MinIOUseSSL:    parseBool(getEnv("MINIO_USE_SSL", "false")),

		MaxUploadSizeBytes: parseInt64(getEnv("MAX_UPLOAD_SIZE_BYTES", "52428800")),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}

func parseBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b
}

func parseInt64(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 52428800
	}
	return n
}
