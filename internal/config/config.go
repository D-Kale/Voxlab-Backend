package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string
	AppPort string
	DB      DatabaseConfig
	Redis   RedisConfig
	MinIO   MinIOConfig
	JWT     JWTConfig
	LiveKit LiveKitConfig
}

type DatabaseConfig struct {
	URL string
}

type RedisConfig struct {
	URL string
}

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type JWTConfig struct {
	Secret string
}

type LiveKitConfig struct {
	Host      string
	APIKey    string
	APISecret string
}

var AppConfig *Config

func LoadConfig() error {
	godotenv.Load(".env")

	AppConfig = &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "3000"),
		DB: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://voxlab:voxlab_pass_2024@localhost:5432/voxlab_db?sslmode=disable"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "localhost:6379"),
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "voxlab_minio_admin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "voxlab_minio_pass_2024"),
			Bucket:    getEnv("MINIO_BUCKET", "voxlab-media"),
			UseSSL:    false,
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "voxlab_jwt_super_secret_key_2024_change_in_production"),
		},
		LiveKit: LiveKitConfig{
			Host:      getEnv("LIVEKIT_HOST", "http://localhost:7880"),
			APIKey:    getEnv("LIVEKIT_API_KEY", "devkey"),
			APISecret: getEnv("LIVEKIT_API_SECRET", "secret"),
		},
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func MustGetConfig() *Config {
	if AppConfig == nil {
		panic("config not initialized — call LoadConfig() first")
	}
	return AppConfig
}
