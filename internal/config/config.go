package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	// Server
	Port        string
	Environment string

	// Database
	MongoURI      string
	MongoDatabase string

	// Redis
	RedisURI      string
	RedisPassword string
	RedisDB       int

	// JWT
	JWTSecret              string
	JWTExpiration          string
	RefreshTokenExpiration string

	// OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURI  string

	// Cloudflare R2
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
	R2PublicURL       string
	R2Endpoint        string

	// Recall.ai
	RecallAPIKey        string
	RecallWebhookSecret string
	RecallBotName       string

	// OpenAI
	OpenAIKey       string
	OpenAIModel     string
	OpenAIMaxTokens int

	// CORS
	AllowedOrigins []string

	// WebRTC
	TURNServerURL  string
	TURNUsername   string
	TURNCredential string
	STUNServerURL  string

	// Rate Limiting
	RateLimitRequests int
	RateLimitWindow   string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		// Server
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENV", "development"),

		// Database
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017/rankedterview"),
		MongoDatabase: getEnv("MONGO_DATABASE", "rankedterview"),

		// Redis
		RedisURI:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		// JWT
		JWTSecret:              getEnv("JWT_SECRET", "your-secret-key-change-this"),
		JWTExpiration:          getEnv("JWT_EXPIRATION", "15m"),
		RefreshTokenExpiration: getEnv("REFRESH_TOKEN_EXPIRATION", "7d"),

		// OAuth
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURI:  getEnv("GOOGLE_REDIRECT_URI", "http://localhost:3000/callback"),
		GitHubClientID:     getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURI:  getEnv("GITHUB_REDIRECT_URI", "http://localhost:3000/callback"),

		// Cloudflare R2
		R2AccountID:       getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:      getEnv("R2_BUCKET_NAME", "rankedterview-recordings"),
		R2PublicURL:       getEnv("R2_PUBLIC_URL", ""),
		R2Endpoint:        getEnv("R2_ENDPOINT", ""),

		// Recall.ai
		RecallAPIKey:        getEnv("RECALL_API_KEY", ""),
		RecallWebhookSecret: getEnv("RECALL_WEBHOOK_SECRET", ""),
		RecallBotName:       getEnv("RECALL_BOT_NAME", "RANKEDterview Recorder"),

		// OpenAI
		OpenAIKey:       getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:     getEnv("OPENAI_MODEL", "gpt-4o"),
		OpenAIMaxTokens: getEnvAsInt("OPENAI_MAX_TOKENS", 2000),

		// CORS
		AllowedOrigins: getEnvAsSlice("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),

		// WebRTC
		TURNServerURL:  getEnv("TURN_SERVER_URL", ""),
		TURNUsername:   getEnv("TURN_USERNAME", ""),
		TURNCredential: getEnv("TURN_CREDENTIAL", ""),
		STUNServerURL:  getEnv("STUN_SERVER_URL", "stun:stun.l.google.com:19302"),

		// Rate Limiting
		RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:   getEnv("RATE_LIMIT_WINDOW", "1m"),
	}
}

// Helper functions

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}
