package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

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
	DatabaseURL string // 完整的資料庫連接字串（生產環境使用）
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
	LogLevel string
}

// LoadConfig 從環境變數載入配置
func LoadConfig() *Config {
	// 先檢查環境類型（從系統環境變數讀取，不從 .env）
	environment := os.Getenv("ENV")
	if environment == "" {
		environment = "development" // 預設為開發環境
	}

	log.Printf("Environment: %s", environment)

	// 根據環境決定是否載入 .env 檔案
	if environment == "development" {
		// 開發環境：載入 .env 檔案
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	}

	config := &Config{
		Environment: environment,
		Port:        getEnv("PORT", "8080"),

		// Sui 設定
		SuiRPCURL:    getEnv("SUI_RPC_URL", "https://fullnode.testnet.sui.io:443"),
		SuiPackageID: getEnv("SUI_PACKAGE_ID", ""),

		// 安全設定
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		SessionTimeout: getEnvAsInt("SESSION_TIMEOUT", 86400), // 24 小時

		// 其他設定
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	// 根據環境決定資料庫配置方式
	if environment == "production" {
		// 生產環境：必須使用 DATABASE_URL
		log.Println("Production mode: Parsing DATABASE_URL...")

		databaseURL := os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			log.Fatal("DATABASE_URL must be set in production environment!")
		}

		config.DatabaseURL = databaseURL
		parsedURL, err := url.Parse(databaseURL)
		if err != nil {
			log.Fatalf("Failed to parse DATABASE_URL: %v", err)
		}

		// 解析使用者名稱和密碼
		config.DBUser = parsedURL.User.Username()
		config.DBPassword, _ = parsedURL.User.Password()

		// 解析主機和埠號
		hostParts := strings.Split(parsedURL.Host, ":")
		config.DBHost = hostParts[0]
		if len(hostParts) > 1 {
			config.DBPort = hostParts[1]
		} else {
			config.DBPort = "5432"
		}

		// 解析資料庫名稱
		config.DBName = strings.TrimPrefix(parsedURL.Path, "/")

		// 解析 SSL 模式
		sslMode := parsedURL.Query().Get("sslmode")
		if sslMode != "" {
			config.DBSSLMode = sslMode
		} else {
			config.DBSSLMode = "require" // 生產環境預設使用 SSL
		}

		log.Printf("Database: %s@%s:%s/%s (SSL: %s)",
			config.DBUser, config.DBHost, config.DBPort, config.DBName, config.DBSSLMode)

	} else {
		// 開發環境：使用個別環境變數（從 .env 讀取）
		log.Println("🔧 Development mode: Using individual DB variables from .env...")

		config.DatabaseURL = ""
		config.DBHost = getEnv("DB_HOST", "localhost")
		config.DBPort = getEnv("DB_PORT", "5432")
		config.DBUser = getEnv("DB_USER", "postgres")
		config.DBPassword = getEnv("DB_PASSWORD", "")
		config.DBName = getEnv("DB_NAME", "bluelink")
		config.DBSSLMode = getEnv("DB_SSL_MODE", "disable")

		log.Printf("Database: %s@%s:%s/%s (SSL: %s)",
			config.DBUser, config.DBHost, config.DBPort, config.DBName, config.DBSSLMode)
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

// func getEnvAsBool(key string, defaultValue bool) bool {
// 	valueStr := os.Getenv(key)
// 	if valueStr == "" {
// 		return defaultValue
// 	}

// 	switch valueStr {
// 	case "true", "1", "yes", "on":
// 		return true
// 	case "false", "0", "no", "off":
// 		return false
// 	default:
// 		log.Printf("Invalid boolean for %s, using default: %t", key, defaultValue)
// 		return defaultValue
// 	}
// }

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
