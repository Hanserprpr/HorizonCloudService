package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config 应用配置
type Config struct {
	App         AppConfig         `json:"app"`
	Server      ServerConfig      `json:"server"`
	Database    DatabaseConfig    `json:"database"`
	Storage     StorageConfig     `json:"storage"`
	JWT         JWTConfig         `json:"jwt"`
	UserService UserServiceConfig `json:"user_service"`
	Quota       QuotaConfig       `json:"quota"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	Debug       bool   `json:"debug"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port               int `json:"port"`
	ReadTimeoutSeconds int `json:"read_timeout_seconds"`
	WriteTimeoutSeconds int `json:"write_timeout_seconds"`
	IdleTimeoutSeconds int `json:"idle_timeout_seconds"`
	MaxHeaderBytes     int `json:"max_header_bytes"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host                     string `json:"host"`
	Port                     int    `json:"port"`
	User                     string `json:"user"`
	Password                 string `json:"password"`
	Name                     string `json:"name"`
	SSLMode                  string `json:"ssl_mode"`
	Timezone                 string `json:"timezone"`
	MaxIdleConns             int    `json:"max_idle_conns"`
	MaxOpenConns             int    `json:"max_open_conns"`
	ConnMaxLifetimeMinutes   int    `json:"conn_max_lifetime_minutes"`
	LogLevel                 string `json:"log_level"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Backend string          `json:"backend"`
	MinIO   MinIOConfig     `json:"minio"`
	S3      S3Config        `json:"s3"`
	Local   LocalConfig     `json:"local"`
}

// MinIOConfig MinIO配置
type MinIOConfig struct {
	Endpoint   string `json:"endpoint"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	BucketName string `json:"bucket_name"`
	Region     string `json:"region"`
	UseSSL     bool   `json:"use_ssl"`
}

// S3Config S3配置
type S3Config struct {
	Region         string `json:"region"`
	AccessKey      string `json:"access_key"`
	SecretKey      string `json:"secret_key"`
	BucketName     string `json:"bucket_name"`
	Endpoint       string `json:"endpoint"`
	ForcePathStyle bool   `json:"force_path_style"`
}

// LocalConfig 本地存储配置
type LocalConfig struct {
	RootPath string `json:"root_path"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret          string `json:"secret"`
	ExpirationHours int    `json:"expiration_hours"`
}

// UserServiceConfig 用户服务配置
type UserServiceConfig struct {
	BaseURL               string `json:"base_url"`
	APIKey                string `json:"api_key"`
	TimeoutSeconds        int    `json:"timeout_seconds"`
	RetryCount            int    `json:"retry_count"`
	RetryIntervalSeconds  int    `json:"retry_interval_seconds"`
	EnableCircuitBreaker  bool   `json:"enable_circuit_breaker"`
}

// QuotaConfig 配额配置
type QuotaConfig struct {
	DefaultStorageQuota    int64   `json:"default_storage_quota"`
	DefaultFileCount       int64   `json:"default_file_count"`
	CheckIntervalMinutes   int     `json:"check_interval_minutes"`
	GraceBuffer           float64 `json:"grace_buffer"`
	EnableWarnings        bool    `json:"enable_warnings"`
	WarningThreshold      float64 `json:"warning_threshold"`
}

// Load 加载配置
func Load() (*Config, error) {
	config := &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "file-service"),
			Version:     getEnv("APP_VERSION", "1.0.0"),
			Environment: getEnv("APP_ENVIRONMENT", "development"),
			Debug:       getEnvBool("APP_DEBUG", true),
		},
		Server: ServerConfig{
			Port:               getEnvInt("SERVER_PORT", 8002),
			ReadTimeoutSeconds: getEnvInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeoutSeconds: getEnvInt("SERVER_WRITE_TIMEOUT", 30),
			IdleTimeoutSeconds: getEnvInt("SERVER_IDLE_TIMEOUT", 120),
			MaxHeaderBytes:     getEnvInt("SERVER_MAX_HEADER_BYTES", 1048576), // 1MB
		},
		Database: DatabaseConfig{
			Host:                   getEnv("DB_HOST", "localhost"),
			Port:                   getEnvInt("DB_PORT", 5432),
			User:                   getEnv("DB_USER", "postgres"),
			Password:               getEnv("DB_PASSWORD", "postgres"),
			Name:                   getEnv("DB_NAME", "file_service"),
			SSLMode:                getEnv("DB_SSL_MODE", "disable"),
			Timezone:               getEnv("DB_TIMEZONE", "UTC"),
			MaxIdleConns:           getEnvInt("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns:           getEnvInt("DB_MAX_OPEN_CONNS", 100),
			ConnMaxLifetimeMinutes: getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 60),
			LogLevel:               getEnv("DB_LOG_LEVEL", "info"),
		},
		Storage: StorageConfig{
			Backend: getEnv("STORAGE_BACKEND", "local"),
			MinIO: MinIOConfig{
				Endpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
				AccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
				SecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
				BucketName: getEnv("MINIO_BUCKET_NAME", "file-service"),
				Region:     getEnv("MINIO_REGION", "us-east-1"),
				UseSSL:     getEnvBool("MINIO_USE_SSL", false),
			},
			S3: S3Config{
				Region:         getEnv("S3_REGION", "us-east-1"),
				AccessKey:      getEnv("S3_ACCESS_KEY", ""),
				SecretKey:      getEnv("S3_SECRET_KEY", ""),
				BucketName:     getEnv("S3_BUCKET_NAME", "file-service"),
				Endpoint:       getEnv("S3_ENDPOINT", ""),
				ForcePathStyle: getEnvBool("S3_FORCE_PATH_STYLE", false),
			},
			Local: LocalConfig{
				RootPath: getEnv("LOCAL_STORAGE_ROOT", "./storage"),
			},
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-development-secret-key"),
			ExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 24),
		},
		UserService: UserServiceConfig{
			BaseURL:              getEnv("USER_SERVICE_BASE_URL", ""),
			APIKey:               getEnv("USER_SERVICE_API_KEY", ""),
			TimeoutSeconds:       getEnvInt("USER_SERVICE_TIMEOUT_SECONDS", 30),
			RetryCount:           getEnvInt("USER_SERVICE_RETRY_COUNT", 3),
			RetryIntervalSeconds: getEnvInt("USER_SERVICE_RETRY_INTERVAL_SECONDS", 5),
			EnableCircuitBreaker: getEnvBool("USER_SERVICE_ENABLE_CIRCUIT_BREAKER", true),
		},
		Quota: QuotaConfig{
			DefaultStorageQuota:  getEnvInt64("QUOTA_DEFAULT_STORAGE_QUOTA", 5*1024*1024*1024), // 5GB
			DefaultFileCount:     getEnvInt64("QUOTA_DEFAULT_FILE_COUNT", 10000),
			CheckIntervalMinutes: getEnvInt("QUOTA_CHECK_INTERVAL_MINUTES", 5),
			GraceBuffer:          getEnvFloat64("QUOTA_GRACE_BUFFER", 0.1),
			EnableWarnings:       getEnvBool("QUOTA_ENABLE_WARNINGS", true),
			WarningThreshold:     getEnvFloat64("QUOTA_WARNING_THRESHOLD", 0.8),
		},
	}

	// 验证配置
	if err := validate(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// validate 验证配置
func validate(config *Config) error {
	// 验证必需的配置
	if config.JWT.Secret == "your-secret-key" && config.App.Environment == "production" {
		return fmt.Errorf("JWT secret must be set in production environment")
	}

	if config.Database.Password == "" && config.App.Environment == "production" {
		return fmt.Errorf("database password must be set")
	}

	// 验证存储后端配置
	switch config.Storage.Backend {
	case "minio":
		if config.Storage.MinIO.AccessKey == "" || config.Storage.MinIO.SecretKey == "" {
			return fmt.Errorf("MinIO access key and secret key must be set")
		}
	case "s3":
		if config.Storage.S3.AccessKey == "" || config.Storage.S3.SecretKey == "" {
			return fmt.Errorf("S3 access key and secret key must be set")
		}
	case "local":
		if config.Storage.Local.RootPath == "" {
			return fmt.Errorf("local storage root path must be set")
		}
	default:
		return fmt.Errorf("unsupported storage backend: %s", config.Storage.Backend)
	}

	return nil
}

// getEnv 获取环境变量字符串值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 获取环境变量整数值
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvInt64 获取环境变量int64值
func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvFloat64 获取环境变量float64值
func getEnvFloat64(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

// getEnvBool 获取环境变量布尔值
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

// GetDuration 获取环境变量时间间隔值
func GetDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// IsDevelopment 检查是否是开发环境
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction 检查是否是生产环境
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsTest 检查是否是测试环境
func (c *Config) IsTest() bool {
	return c.App.Environment == "test"
}