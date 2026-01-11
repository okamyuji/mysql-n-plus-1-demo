package config

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// Config - アプリケーション設定
type Config struct {
	DBHost      string
	DBPort      int
	DBName      string
	DBUser      string
	DBPassword  string
	DBCharset   string
	DBParseTime bool
}

// LoadConfig - 設定を読み込む
func LoadConfig() (*Config, error) {
	// .envファイルを読み込む（存在する場合）
	_ = godotenv.Load()

	config := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBName:     getEnv("DB_NAME", "demo_db"),
		DBUser:     getEnv("DB_USER", "demo_user"),
		DBPassword: getEnv("DB_PASSWORD", "demo_password"),
		DBCharset:  getEnv("DB_CHARSET", "utf8mb4"),
	}

	// DBポート番号の解析
	portStr := getEnv("DB_PORT", "3306")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}
	config.DBPort = port

	// parseTime設定の解析
	parseTimeStr := getEnv("DB_PARSE_TIME", "true")
	parseTime, err := strconv.ParseBool(parseTimeStr)
	if err != nil {
		parseTime = true
	}
	config.DBParseTime = parseTime

	// 必須項目のチェック
	if config.DBUser == "" {
		return nil, fmt.Errorf("DB_USER is required")
	}
	if config.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	return config, nil
}

// ConnectDatabase - データベースに接続する
func ConnectDatabase(config *Config) (*sql.DB, error) {
	// MySQL DSN (Data Source Name) の構築
	// format: user:password@tcp(host:port)/dbname?param=value
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=Local",
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBPort,
		config.DBName,
		config.DBCharset,
		config.DBParseTime,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 接続プールの設定
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// getEnv - 環境変数を取得（デフォルト値付き）
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
