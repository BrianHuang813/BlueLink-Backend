package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config 應用程式配置
type Config struct {
	// 環境設定
	Environment string // "development" | "production"
	Port        string // "8080"

	// Sui 區塊鏈設定
	SuiRPCURL    string
	SuiPackageID string // 合約 Package ID

	// 資料庫設定
	DatabaseURL string // "postgres://..."
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string

	// JWT/Session 設定
	JWTSecret      string
	SessionTimeout int // 秒

	// 其他設定
	LogLevel      string
	EnableSwagger bool
}

// LoadConfig 從環境變數載入配置
func LoadConfig() *Config {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found.")
	}

	config := &Config{
		// 環境設定
		Environment: getEnv("ENV", "development"),
		Port:        getEnv("PORT", ":8080"),

		// Sui 設定
		SuiRPCURL:    getEnv("SUI_RPC_URL", "https://fullnode.testnet.sui.io:443"),
		SuiPackageID: getEnv("SUI_PACKAGE_ID", ""), // 合約部署後設定

		// 資料庫設定
		DatabaseURL: getEnv("DATABASE_URL", ""),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", ""),
		DBName:      getEnv("DB_NAME", "bluelink"),
		DBSSLMode:   getEnv("DB_SSL_MODE", "disable"), // AWS RDS 使用 "require"

		// 安全設定
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		SessionTimeout: getEnvAsInt("SESSION_TIMEOUT", 86400), // 24 小時

		// 其他設定
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		EnableSwagger: getEnvAsBool("ENABLE_SWAGGER", true),
	}

	// 驗證必要的配置
	config.Validate()

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	var value int
	if _, err := fmt.Sscanf(valueStr, "%d", &value); err != nil {
		log.Printf("Invalid integer for %s, using default: %d", key, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	switch valueStr {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		log.Printf("Invalid boolean for %s, using default: %t", key, defaultValue)
		return defaultValue
	}
}

func (c *Config) Validate() {
	if c.Environment == "production" {
		// 生產環境環境變數設定
		if c.JWTSecret == "your-secret-key-change-in-production" {
			log.Fatal("JWT_SECRET must be set in production!")
		}
		if c.DBPassword == "" {
			log.Fatal("DB_PASSWORD must be set in production!")
		}
	}

	log.Printf("Configuration loaded successfully")
}
