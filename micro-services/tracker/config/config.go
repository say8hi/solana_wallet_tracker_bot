package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
    SolanaRPC struct {
        WebsocketURL    string
        HTTPUrl        string
        Commitment     string
        RetryInterval  time.Duration
        MaxRetries     int
    }

    Redis struct {
        HOST                 string
        PORT                 string
        PASSWORD                 string
        TransactionsChannel string
        CommandsChannel     string
        Password           string
        Database           int
        PoolSize           int
    }

    App struct {
        Name            string
        Environment     string
        LogLevel        string
        ShutdownTimeout time.Duration
    }

    Monitoring struct {
        EnableMetrics    bool
        MetricsPort     int
        HealthCheckPort int
    }
}

func LoadConfig() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        
    }

    cfg := &Config{}

    // Solana
    cfg.SolanaRPC.WebsocketURL = getEnvOrDefault("SOLANA_WS_URL", "wss://mainnet.helius-rpc.com/?api-key=b5173fe8-f2ff-4ace-b3f8-ec0a74465231")
    cfg.SolanaRPC.HTTPUrl = getEnvOrDefault("SOLANA_HTTP_URL", "wss://mainnet.helius-rpc.com/?api-key=b5173fe8-f2ff-4ace-b3f8-ec0a74465231")
    cfg.SolanaRPC.Commitment = getEnvOrDefault("SOLANA_COMMITMENT", "processed")
    cfg.SolanaRPC.RetryInterval = getDurationOrDefault("SOLANA_RETRY_INTERVAL", 5*time.Second)
    cfg.SolanaRPC.MaxRetries = getIntOrDefault("SOLANA_MAX_RETRIES", 5)

    // Redis 
    cfg.Redis.HOST = getEnvOrDefault("REDIS_HOST", "tracker_redis")
    cfg.Redis.PORT = getEnvOrDefault("REDIS_PORT", "6379")
    cfg.Redis.PASSWORD = getEnvOrDefault("REDIS_PASSWORD", "someredispass")
    cfg.Redis.TransactionsChannel = getEnvOrDefault("REDIS_TX_CHANNEL", "solana_transactions")
    cfg.Redis.CommandsChannel = getEnvOrDefault("REDIS_CMD_CHANNEL", "wallet_commands")
    cfg.Redis.Password = getEnvOrDefault("REDIS_PASSWORD", "")
    cfg.Redis.Database = getIntOrDefault("REDIS_DB", 0)
    cfg.Redis.PoolSize = getIntOrDefault("REDIS_POOL_SIZE", 10)

    // App 
    cfg.App.Name = getEnvOrDefault("APP_NAME", "solana-tracker")
    cfg.App.Environment = getEnvOrDefault("APP_ENV", "development")
    cfg.App.LogLevel = getEnvOrDefault("LOG_LEVEL", "info")
    cfg.App.ShutdownTimeout = getDurationOrDefault("SHUTDOWN_TIMEOUT", 30*time.Second)

    // Monitoring 
    cfg.Monitoring.EnableMetrics = getBoolOrDefault("ENABLE_METRICS", true)
    cfg.Monitoring.MetricsPort = getIntOrDefault("METRICS_PORT", 9090)
    cfg.Monitoring.HealthCheckPort = getIntOrDefault("HEALTH_PORT", 8080)

    return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getIntOrDefault(key string, defaultValue int) int {
    strValue := os.Getenv(key)
    if strValue == "" {
        return defaultValue
    }
    
    value, err := strconv.Atoi(strValue)
    if err != nil {
        return defaultValue
    }
    
    return value
}

func getBoolOrDefault(key string, defaultValue bool) bool {
    strValue := os.Getenv(key)
    if strValue == "" {
        return defaultValue
    }
    
    value, err := strconv.ParseBool(strValue)
    if err != nil {
        return defaultValue
    }
    
    return value
}

func getDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
    strValue := os.Getenv(key)
    if strValue == "" {
        return defaultValue
    }
    
    value, err := time.ParseDuration(strValue)
    if err != nil {
        return defaultValue
    }
    
    return value
}
