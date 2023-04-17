package config

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

type Config struct {
	Auth0ClientId             string
	Auth0ClientSecret         string
	Auth0Domain               string
	GoogleStorageSync         bool
	GoogleStorageSyncBucket   string
	GoogleStorageSyncInterval int
	LogLevel                  log.Level
	Port                      int
	SiteDir                   string
	Tls                       bool
	WebFolder                 string
	WebFendrHost              string
	WebFendrMode              string
}

func (c *Config) HttpProtocol() string {
	if c.Tls {
		return "https"
	}
	return "http"
}

func GetEnv(key string) string {
	value, exist := os.LookupEnv(key)
	if !exist {
		panic(fmt.Sprintf("Missing env var: %s", key))
	}
	return value
}

func GetEnvFallback(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func GetEnvAtoi(key string, fallback int) int {
	value := os.Getenv(key)
	i, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return i
}

func GetEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	b, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return b
}

func PrepareConfig() *Config {
	err := godotenv.Load(".env")
	if err != nil {
		// Ignore
	}

	levelStr := GetEnvFallback("LOG_LEVEL", "info")
	level := log.InfoLevel
	if levelStr == "fatal" {
		level = log.FatalLevel
	} else if levelStr == "error" {
		level = log.ErrorLevel
	} else if levelStr == "debug" {
		level = log.DebugLevel
	} else if levelStr == "trace" {
		level = log.TraceLevel
	}

	webFendrModeStr := GetEnvFallback("WEBFENDR_MODE", "release")
	webFendrMode := gin.ReleaseMode
	if webFendrModeStr == "debug" {
		webFendrMode = gin.DebugMode
	}

	log.Info("WebFendr mode: ", webFendrMode)

	return &Config{
		Auth0ClientId:             GetEnv("AUTH0_CLIENT_ID"),
		Auth0ClientSecret:         GetEnv("AUTH0_CLIENT_SECRET"),
		Auth0Domain:               GetEnv("AUTH0_DOMAIN"),
		GoogleStorageSync:         GetEnvBool("GOOGLE_STORAGE_SYNC", false),
		GoogleStorageSyncBucket:   GetEnvFallback("GOOGLE_STORAGE_SYNC_BUCKET", ""),
		GoogleStorageSyncInterval: GetEnvAtoi("GOOGLE_STORAGE_SYNC_INTERVAL", 300),
		LogLevel:                  level,
		Port:                      GetEnvAtoi("PORT", 3000),
		SiteDir:                   GetEnv("SITE_DIR"),
		Tls:                       GetEnvBool("TLS", true),
		WebFolder:                 GetEnvFallback("WEB_FOLDER", "."),
		WebFendrHost:              GetEnv("WEBFENDR_HOST"),
		WebFendrMode:              webFendrMode,
	}
}
