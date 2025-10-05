package config

func LoadConfig() *Config {
	// 這裡可以從檔案、環境變數或其他來源載入設定
	return &Config{
		Environment: "development", // 或 "production"
	}
}

type Config struct {
	Environment string
}
