package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	RateLimit   RateLimitConfig   `mapstructure:"rate_limit"`
	Scheduler   SchedulerConfig   `mapstructure:"scheduler"`
	JWT         JWTConfig         `mapstructure:"jwt"`
	Log         LogConfig         `mapstructure:"log"`
	Telemetry   TelemetryConfig   `mapstructure:"telemetry"`
	Integration IntegrationConfig `mapstructure:"integration"`
	Menu        MenuConfig        `mapstructure:"menu"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Name         string `mapstructure:"name"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type RateLimitConfig struct {
	Enabled              bool                        `mapstructure:"enabled"`
	DefaultMaxRequests   int                         `mapstructure:"default_max_requests"`
	DefaultWindowSeconds int                         `mapstructure:"default_window_seconds"`
	UseRedis             bool                        `mapstructure:"use_redis"`
	RouteOverrides       map[string]RouteLimitConfig `mapstructure:"route_overrides"`
}

type RouteLimitConfig struct {
	Enabled       *bool `mapstructure:"enabled"`
	MaxRequests   int   `mapstructure:"max_requests"`
	WindowSeconds int   `mapstructure:"window_seconds"`
}

type SchedulerConfig struct {
	SampleJobEnabled bool `mapstructure:"sample_job_enabled"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type TelemetryConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
}

type IntegrationConfig struct {
	Calcite   GRPCIntegrationConfig `mapstructure:"calcite"`
	Seatunnel GRPCIntegrationConfig `mapstructure:"seatunnel"`
}

type GRPCIntegrationConfig struct {
	Address    string `mapstructure:"address"`
	TimeoutSec int    `mapstructure:"timeout_sec"`
	MaxRetries int    `mapstructure:"max_retries"`
}

// MenuConfig 菜单配置
type MenuConfig struct {
	HardcodedFallback bool `mapstructure:"hardcoded_fallback"`
}

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./configs"
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath("./configs")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := bindEnvKeys(); err != nil {
		return nil, err
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	applyDefaults(&config)

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func bindEnvKeys() error {
	keys := map[string]string{
		"server.port":                       "SERVER_PORT",
		"server.mode":                       "SERVER_MODE",
		"database.host":                     "DATABASE_HOST",
		"database.port":                     "DATABASE_PORT",
		"database.name":                     "DATABASE_NAME",
		"database.user":                     "DATABASE_USER",
		"database.password":                 "DATABASE_PASSWORD",
		"database.max_open_conns":           "DATABASE_MAX_OPEN_CONNS",
		"database.max_idle_conns":           "DATABASE_MAX_IDLE_CONNS",
		"redis.host":                        "REDIS_HOST",
		"redis.port":                        "REDIS_PORT",
		"redis.password":                    "REDIS_PASSWORD",
		"redis.db":                          "REDIS_DB",
		"redis.pool_size":                   "REDIS_POOL_SIZE",
		"rate_limit.enabled":                "RATE_LIMIT_ENABLED",
		"rate_limit.default_max_requests":   "RATE_LIMIT_DEFAULT_MAX_REQUESTS",
		"rate_limit.default_window_seconds": "RATE_LIMIT_DEFAULT_WINDOW_SECONDS",
		"rate_limit.use_redis":              "RATE_LIMIT_USE_REDIS",
		"scheduler.sample_job_enabled":      "SCHEDULER_SAMPLE_JOB_ENABLED",
		"jwt.secret":                        "JWT_SECRET",
		"jwt.expire":                        "JWT_EXPIRE",
		"log.level":                         "LOG_LEVEL",
		"log.format":                        "LOG_FORMAT",
		"telemetry.enabled":                 "TELEMETRY_ENABLED",
		"telemetry.endpoint":                "TELEMETRY_ENDPOINT",
		"integration.calcite.address":       "CALCITE_GRPC_ADDR",
		"integration.calcite.timeout_sec":   "CALCITE_GRPC_TIMEOUT_SEC",
		"integration.calcite.max_retries":   "CALCITE_GRPC_MAX_RETRIES",
		"integration.seatunnel.address":     "SEATUNNEL_GRPC_ADDR",
		"integration.seatunnel.timeout_sec": "SEATUNNEL_GRPC_TIMEOUT_SEC",
		"integration.seatunnel.max_retries": "SEATUNNEL_GRPC_MAX_RETRIES",
		"menu.hardcoded_fallback":           "MENU_HARDCODED_FALLBACK",
	}

	for key, envName := range keys {
		if err := viper.BindEnv(key, envName); err != nil {
			return fmt.Errorf("failed to bind env %s for key %s: %w", envName, key, err)
		}
	}

	return nil
}

func applyDefaults(config *Config) {
	if config == nil {
		return
	}

	if config.RateLimit.DefaultMaxRequests <= 0 {
		config.RateLimit.DefaultMaxRequests = 100
	}
	if config.RateLimit.DefaultWindowSeconds <= 0 {
		config.RateLimit.DefaultWindowSeconds = 60
	}
	if !viper.IsSet("rate_limit.use_redis") && os.Getenv("RATE_LIMIT_USE_REDIS") == "" {
		config.RateLimit.UseRedis = true
	}

	for key, override := range config.RateLimit.RouteOverrides {
		if override.MaxRequests <= 0 {
			override.MaxRequests = config.RateLimit.DefaultMaxRequests
		}
		if override.WindowSeconds <= 0 {
			override.WindowSeconds = config.RateLimit.DefaultWindowSeconds
		}
		config.RateLimit.RouteOverrides[key] = override
	}
}

func validateConfig(config *Config) error {
	if config.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}
	if config.Database.Name == "" {
		return fmt.Errorf("database.name is required")
	}
	if config.Redis.Host == "" {
		return fmt.Errorf("redis.host is required")
	}
	if config.JWT.Secret == "" {
		return fmt.Errorf("jwt.secret must be set")
	}
	if config.RateLimit.DefaultMaxRequests <= 0 {
		return fmt.Errorf("rate_limit.default_max_requests must be greater than 0")
	}
	if config.RateLimit.DefaultWindowSeconds <= 0 {
		return fmt.Errorf("rate_limit.default_window_seconds must be greater than 0")
	}
	for routeName, override := range config.RateLimit.RouteOverrides {
		if override.MaxRequests <= 0 {
			return fmt.Errorf("rate_limit.route_overrides.%s.max_requests must be greater than 0", routeName)
		}
		if override.WindowSeconds <= 0 {
			return fmt.Errorf("rate_limit.route_overrides.%s.window_seconds must be greater than 0", routeName)
		}
	}
	return nil
}
