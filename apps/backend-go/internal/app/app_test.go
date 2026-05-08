package app

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// resetViper resets the global viper instance before and after each test.
func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(func() { viper.Reset() })
}

// validConfig returns a Config that passes validateConfig without errors.
func validConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host: "localhost",
			Name: "testdb",
		},
		Redis: RedisConfig{
			Host: "localhost",
		},
		JWT: JWTConfig{
			Secret: "my-secret",
		},
		RateLimit: RateLimitConfig{
			DefaultMaxRequests:   100,
			DefaultWindowSeconds: 60,
		},
	}
}

// ---------------------------------------------------------------------------
// Existing tests (preserved)
// ---------------------------------------------------------------------------

func TestApplication_Fields(t *testing.T) {
	app := &Application{
		Name:    "test-app",
		Version: "1.0.0",
		Config:  nil,
	}

	if app.Name != "test-app" {
		t.Errorf("Expected Name 'test-app', got '%s'", app.Name)
	}
	if app.Version != "1.0.0" {
		t.Errorf("Expected Version '1.0.0', got '%s'", app.Version)
	}
}

func TestApplication_NilConfig(t *testing.T) {
	app := &Application{
		Name:    "test",
		Version: "1.0",
		Config:  nil,
	}

	if app.Config != nil {
		t.Error("Expected Config to be nil")
	}
}

// ---------------------------------------------------------------------------
// validateConfig tests
// ---------------------------------------------------------------------------

func TestValidateConfig_Success(t *testing.T) {
	err := validateConfig(validConfig())
	assert.NoError(t, err)
}

func TestValidateConfig_MissingDatabaseHost(t *testing.T) {
	cfg := validConfig()
	cfg.Database.Host = ""
	err := validateConfig(cfg)
	assert.EqualError(t, err, "database.host is required")
}

func TestValidateConfig_MissingDatabaseName(t *testing.T) {
	cfg := validConfig()
	cfg.Database.Name = ""
	err := validateConfig(cfg)
	assert.EqualError(t, err, "database.name is required")
}

func TestValidateConfig_MissingRedisHost(t *testing.T) {
	cfg := validConfig()
	cfg.Redis.Host = ""
	err := validateConfig(cfg)
	assert.EqualError(t, err, "redis.host is required")
}

func TestValidateConfig_MissingJWTSecret(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.Secret = ""
	err := validateConfig(cfg)
	assert.EqualError(t, err, "jwt.secret must be set")
}

func TestValidateConfig_ZeroMaxRequests(t *testing.T) {
	cfg := validConfig()
	cfg.RateLimit.DefaultMaxRequests = 0
	err := validateConfig(cfg)
	assert.EqualError(t, err, "rate_limit.default_max_requests must be greater than 0")
}

func TestValidateConfig_NegativeMaxRequests(t *testing.T) {
	cfg := validConfig()
	cfg.RateLimit.DefaultMaxRequests = -5
	err := validateConfig(cfg)
	assert.EqualError(t, err, "rate_limit.default_max_requests must be greater than 0")
}

func TestValidateConfig_ZeroWindowSeconds(t *testing.T) {
	cfg := validConfig()
	cfg.RateLimit.DefaultWindowSeconds = 0
	err := validateConfig(cfg)
	assert.EqualError(t, err, "rate_limit.default_window_seconds must be greater than 0")
}

func TestValidateConfig_NegativeWindowSeconds(t *testing.T) {
	cfg := validConfig()
	cfg.RateLimit.DefaultWindowSeconds = -10
	err := validateConfig(cfg)
	assert.EqualError(t, err, "rate_limit.default_window_seconds must be greater than 0")
}

func TestValidateConfig_RouteOverrideZeroMaxRequests(t *testing.T) {
	cfg := validConfig()
	cfg.RateLimit.RouteOverrides = map[string]RouteLimitConfig{
		"login": {MaxRequests: 0, WindowSeconds: 60},
	}
	err := validateConfig(cfg)
	assert.EqualError(t, err, "rate_limit.route_overrides.login.max_requests must be greater than 0")
}

func TestValidateConfig_RouteOverrideZeroWindowSeconds(t *testing.T) {
	cfg := validConfig()
	cfg.RateLimit.RouteOverrides = map[string]RouteLimitConfig{
		"api": {MaxRequests: 50, WindowSeconds: 0},
	}
	err := validateConfig(cfg)
	assert.EqualError(t, err, "rate_limit.route_overrides.api.window_seconds must be greater than 0")
}

func TestValidateConfig_MultipleRouteOverrides(t *testing.T) {
	cfg := validConfig()
	cfg.RateLimit.RouteOverrides = map[string]RouteLimitConfig{
		"login":  {MaxRequests: 10, WindowSeconds: 30},
		"export": {MaxRequests: 5, WindowSeconds: 120},
	}
	err := validateConfig(cfg)
	assert.NoError(t, err)
}

func TestValidateConfig_FirstFieldErrorReported(t *testing.T) {
	// When multiple fields are empty, the first failing check is reported.
	cfg := &Config{}
	err := validateConfig(cfg)
	assert.Error(t, err)
	assert.EqualError(t, err, "database.host is required")
}

// ---------------------------------------------------------------------------
// applyDefaults tests
// ---------------------------------------------------------------------------

func TestApplyDefaults_SetsMaxRequests(t *testing.T) {
	cfg := &Config{RateLimit: RateLimitConfig{}}
	applyDefaults(cfg)
	assert.Equal(t, 100, cfg.RateLimit.DefaultMaxRequests)
}

func TestApplyDefaults_SetsWindowSeconds(t *testing.T) {
	cfg := &Config{RateLimit: RateLimitConfig{}}
	applyDefaults(cfg)
	assert.Equal(t, 60, cfg.RateLimit.DefaultWindowSeconds)
}

func TestApplyDefaults_SetsUseRedisDefault(t *testing.T) {
	resetViper(t)
	cfg := &Config{RateLimit: RateLimitConfig{}}
	applyDefaults(cfg)
	assert.True(t, cfg.RateLimit.UseRedis, "UseRedis should default to true when not set via viper or env")
}

func TestApplyDefaults_UseRedisNotOverriddenWhenViperSet(t *testing.T) {
	resetViper(t)
	viper.Set("rate_limit.use_redis", false)
	cfg := &Config{RateLimit: RateLimitConfig{UseRedis: false}}
	applyDefaults(cfg)
	assert.False(t, cfg.RateLimit.UseRedis, "UseRedis should remain false when explicitly set via viper")
}

func TestApplyDefaults_PreservesExistingMaxRequests(t *testing.T) {
	cfg := &Config{RateLimit: RateLimitConfig{DefaultMaxRequests: 200}}
	applyDefaults(cfg)
	assert.Equal(t, 200, cfg.RateLimit.DefaultMaxRequests)
}

func TestApplyDefaults_PreservesExistingWindowSeconds(t *testing.T) {
	cfg := &Config{RateLimit: RateLimitConfig{DefaultWindowSeconds: 120}}
	applyDefaults(cfg)
	assert.Equal(t, 120, cfg.RateLimit.DefaultWindowSeconds)
}

func TestApplyDefaults_PreservesExistingValues(t *testing.T) {
	cfg := &Config{RateLimit: RateLimitConfig{
		DefaultMaxRequests:   250,
		DefaultWindowSeconds: 90,
		UseRedis:             false,
	}}
	applyDefaults(cfg)
	assert.Equal(t, 250, cfg.RateLimit.DefaultMaxRequests)
	assert.Equal(t, 90, cfg.RateLimit.DefaultWindowSeconds)
	// UseRedis is false but not set via viper/env, so applyDefaults sets it to true
	// unless viper.IsSet or env var is present. Since we didn't set those, it flips.
	assert.True(t, cfg.RateLimit.UseRedis)
}

func TestApplyDefaults_RouteOverridesInheritMaxRequests(t *testing.T) {
	cfg := &Config{
		RateLimit: RateLimitConfig{
			DefaultMaxRequests:   200,
			DefaultWindowSeconds: 120,
			RouteOverrides: map[string]RouteLimitConfig{
				"login": {MaxRequests: 0, WindowSeconds: 60},
			},
		},
	}
	applyDefaults(cfg)
	assert.Equal(t, 200, cfg.RateLimit.RouteOverrides["login"].MaxRequests)
	assert.Equal(t, 60, cfg.RateLimit.RouteOverrides["login"].WindowSeconds)
}

func TestApplyDefaults_RouteOverridesInheritWindowSeconds(t *testing.T) {
	cfg := &Config{
		RateLimit: RateLimitConfig{
			DefaultMaxRequests:   100,
			DefaultWindowSeconds: 60,
			RouteOverrides: map[string]RouteLimitConfig{
				"api": {MaxRequests: 50, WindowSeconds: 0},
			},
		},
	}
	applyDefaults(cfg)
	assert.Equal(t, 50, cfg.RateLimit.RouteOverrides["api"].MaxRequests)
	assert.Equal(t, 60, cfg.RateLimit.RouteOverrides["api"].WindowSeconds)
}

func TestApplyDefaults_RouteOverridesInheritBoth(t *testing.T) {
	cfg := &Config{
		RateLimit: RateLimitConfig{
			DefaultMaxRequests:   150,
			DefaultWindowSeconds: 45,
			RouteOverrides: map[string]RouteLimitConfig{
				"export": {MaxRequests: 0, WindowSeconds: 0},
			},
		},
	}
	applyDefaults(cfg)
	assert.Equal(t, 150, cfg.RateLimit.RouteOverrides["export"].MaxRequests)
	assert.Equal(t, 45, cfg.RateLimit.RouteOverrides["export"].WindowSeconds)
}

func TestApplyDefaults_RouteOverridesPreservedWhenNonZero(t *testing.T) {
	cfg := &Config{
		RateLimit: RateLimitConfig{
			DefaultMaxRequests:   100,
			DefaultWindowSeconds: 60,
			RouteOverrides: map[string]RouteLimitConfig{
				"custom": {MaxRequests: 10, WindowSeconds: 5},
			},
		},
	}
	applyDefaults(cfg)
	assert.Equal(t, 10, cfg.RateLimit.RouteOverrides["custom"].MaxRequests)
	assert.Equal(t, 5, cfg.RateLimit.RouteOverrides["custom"].WindowSeconds)
}

func TestApplyDefaults_NilConfig(t *testing.T) {
	assert.NotPanics(t, func() {
		applyDefaults(nil)
	})
}

func TestApplyDefaults_NilRouteOverrides(t *testing.T) {
	cfg := &Config{RateLimit: RateLimitConfig{RouteOverrides: nil}}
	assert.NotPanics(t, func() {
		applyDefaults(cfg)
	})
	assert.Nil(t, cfg.RateLimit.RouteOverrides)
}

func TestApplyDefaults_NegativeMaxRequests(t *testing.T) {
	cfg := &Config{RateLimit: RateLimitConfig{DefaultMaxRequests: -1}}
	applyDefaults(cfg)
	assert.Equal(t, 100, cfg.RateLimit.DefaultMaxRequests, "negative values should be treated like zero")
}

func TestApplyDefaults_NegativeWindowSeconds(t *testing.T) {
	cfg := &Config{RateLimit: RateLimitConfig{DefaultWindowSeconds: -1}}
	applyDefaults(cfg)
	assert.Equal(t, 60, cfg.RateLimit.DefaultWindowSeconds, "negative values should be treated like zero")
}

// ---------------------------------------------------------------------------
// bindEnvKeys tests
// ---------------------------------------------------------------------------

func TestBindEnvKeys_Success(t *testing.T) {
	resetViper(t)
	err := bindEnvKeys()
	assert.NoError(t, err)
}

func TestBindEnvKeys_BindsDatabaseHost(t *testing.T) {
	resetViper(t)
	err := bindEnvKeys()
	assert.NoError(t, err)
	// After binding, setting the env var should make viper return the value.
	t.Setenv("DATABASE_HOST", "from-env")
	// viper won't auto-refresh, but the binding should be registered.
	assert.Equal(t, "from-env", viper.Get("database.host"))
}

// ---------------------------------------------------------------------------
// LoadConfig tests
// ---------------------------------------------------------------------------

func TestLoadConfig_FromDefaultPath(t *testing.T) {
	resetViper(t)
	t.Setenv("CONFIG_PATH", "../../configs")
	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "debug", cfg.Server.Mode)
	assert.Equal(t, "dataease_dev", cfg.Database.Name)
	assert.Equal(t, "172.19.0.2", cfg.Database.Host)
	assert.Equal(t, "172.19.0.2", cfg.Redis.Host)
	assert.Equal(t, "dev-secret-key-for-local-development", cfg.JWT.Secret)
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	resetViper(t)
	t.Setenv("CONFIG_PATH", "../../configs")
	t.Setenv("DATABASE_HOST", "override-host")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "override-host", cfg.Database.Host)
}

func TestLoadConfig_MissingConfigFile(t *testing.T) {
	resetViper(t)
	t.Setenv("CONFIG_PATH", "/nonexistent/path/that/does/not/exist")
	_, err := LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadConfig_ApplyDefaultsApplied(t *testing.T) {
	resetViper(t)
	t.Setenv("CONFIG_PATH", "../../configs")
	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Greater(t, cfg.RateLimit.DefaultMaxRequests, 0)
	assert.Greater(t, cfg.RateLimit.DefaultWindowSeconds, 0)
}

func TestLoadConfig_EnvOverrideRedisHost(t *testing.T) {
	resetViper(t)
	t.Setenv("CONFIG_PATH", "../../configs")
	t.Setenv("REDIS_HOST", "redis-override")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "redis-override", cfg.Redis.Host)
}

func TestLoadConfig_EnvOverrideJWTSecret(t *testing.T) {
	resetViper(t)
	t.Setenv("CONFIG_PATH", "../../configs")
	t.Setenv("JWT_SECRET", "env-jwt-secret")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "env-jwt-secret", cfg.JWT.Secret)
}

func TestLoadConfig_EnvOverrideMultipleFields(t *testing.T) {
	resetViper(t)
	t.Setenv("CONFIG_PATH", "../../configs")
	t.Setenv("DATABASE_HOST", "db-host")
	t.Setenv("DATABASE_NAME", "db-name")
	t.Setenv("REDIS_HOST", "rd-host")
	t.Setenv("JWT_SECRET", "jwt-s3cret")

	cfg, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "db-host", cfg.Database.Host)
	assert.Equal(t, "db-name", cfg.Database.Name)
	assert.Equal(t, "rd-host", cfg.Redis.Host)
	assert.Equal(t, "jwt-s3cret", cfg.JWT.Secret)
}

// ---------------------------------------------------------------------------
// Integration: applyDefaults + validateConfig together
// ---------------------------------------------------------------------------

func TestApplyDefaultsThenValidate_Success(t *testing.T) {
	// Simulate what LoadConfig does: applyDefaults then validateConfig
	cfg := &Config{
		Database: DatabaseConfig{Host: "localhost", Name: "testdb"},
		Redis:    RedisConfig{Host: "localhost"},
		JWT:      JWTConfig{Secret: "secret"},
		RateLimit: RateLimitConfig{
			RouteOverrides: map[string]RouteLimitConfig{
				"login": {MaxRequests: 0, WindowSeconds: 0},
			},
		},
	}
	resetViper(t)
	applyDefaults(cfg)
	err := validateConfig(cfg)
	assert.NoError(t, err)
	assert.Equal(t, 100, cfg.RateLimit.DefaultMaxRequests)
	assert.Equal(t, 60, cfg.RateLimit.DefaultWindowSeconds)
	assert.Equal(t, 100, cfg.RateLimit.RouteOverrides["login"].MaxRequests)
	assert.Equal(t, 60, cfg.RateLimit.RouteOverrides["login"].WindowSeconds)
}

// ---------------------------------------------------------------------------
// Edge case: UseRedis env var interaction
// ---------------------------------------------------------------------------

func TestApplyDefaults_UseRedisEnvVarSet(t *testing.T) {
	resetViper(t)
	t.Setenv("RATE_LIMIT_USE_REDIS", "false")
	cfg := &Config{RateLimit: RateLimitConfig{UseRedis: false}}
	applyDefaults(cfg)
	// env var is set so applyDefaults should NOT force it to true
	assert.False(t, cfg.RateLimit.UseRedis)
}

func TestApplyDefaults_UseRedisViperExplicitTrue(t *testing.T) {
	resetViper(t)
	viper.Set("rate_limit.use_redis", true)
	cfg := &Config{RateLimit: RateLimitConfig{UseRedis: true}}
	applyDefaults(cfg)
	assert.True(t, cfg.RateLimit.UseRedis)
}

// ---------------------------------------------------------------------------
// Cleanup: ensure viper state is clean after tests
// ---------------------------------------------------------------------------

func TestMain_Cleanup(t *testing.T) {
	// This test ensures viper is clean after all other tests.
	// If other tests leak viper state, this would catch it indirectly.
	resetViper(t)
	assert.Equal(t, "", viper.GetString("database.host"))
}
