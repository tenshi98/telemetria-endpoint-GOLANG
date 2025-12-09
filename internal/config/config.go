package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config -> almacena toda la configuración de la aplicación
type Config struct {
	Server    ServerConfig
	MySQL     MySQLConfig
	Redis     RedisConfig
	MQTT      MQTTConfig
	RateLimit RateLimitConfig
	Logging   LoggingConfig
}

// Configuración de Base de Datos MySQL
type MySQLConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Configuración de Redis
type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
	CacheTTL     time.Duration
}

// Configuración de MQTT
type MQTTConfig struct {
	Enabled      bool
	BrokerURL    string
	ClientID     string
	Username     string
	Password     string
	Topic        string
	QoS          byte
	CleanSession bool
}

// Configuración de HTTP Server
type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// Configuración de Rate Limiting
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	RequestDelay      time.Duration
}

// Configuración de Logging
type LoggingConfig struct {
	LogDir          string
	AppLogFile      string
	InvalidLogFile  string
	DeviceLogDir    string
}

// Load -> carga la configuración desde variables de entorno
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		MySQL: MySQLConfig{
			Host:            getEnv("MYSQL_HOST", "localhost"),
			Port:            getEnv("MYSQL_PORT", "3306"),
			User:            getEnv("MYSQL_USER", "root"),
			Password:        getEnv("MYSQL_PASSWORD", ""),
			Database:        getEnv("MYSQL_DATABASE", "telemetria"),
			MaxOpenConns:    getIntEnv("MYSQL_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("MYSQL_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("MYSQL_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnv("REDIS_PORT", "6379"),
			Password:     getEnv("REDIS_PASSWORD", ""),
			DB:           getIntEnv("REDIS_DB", 0),
			MaxRetries:   getIntEnv("REDIS_MAX_RETRIES", 3),
			PoolSize:     getIntEnv("REDIS_POOL_SIZE", 10),
			MinIdleConns: getIntEnv("REDIS_MIN_IDLE_CONNS", 2),
			CacheTTL:     getDurationEnv("REDIS_CACHE_TTL", 24*time.Hour),
		},
		MQTT: MQTTConfig{
			Enabled:      getBoolEnv("MQTT_ENABLED", false),
			BrokerURL:    getEnv("MQTT_BROKER_URL", "tcp://localhost:1883"),
			ClientID:     getEnv("MQTT_CLIENT_ID", "telemetry-endpoint"),
			Username:     getEnv("MQTT_USERNAME", ""),
			Password:     getEnv("MQTT_PASSWORD", ""),
			Topic:        getEnv("MQTT_TOPIC", "telemetry/data"),
			QoS:          byte(getIntEnv("MQTT_QOS", 1)),
			CleanSession: getBoolEnv("MQTT_CLEAN_SESSION", true),
		},
		RateLimit: RateLimitConfig{
			RequestsPerSecond: getFloat64Env("RATE_LIMIT_RPS", 100.0),
			BurstSize:         getIntEnv("RATE_LIMIT_BURST", 200),
			RequestDelay:      getDurationEnv("REQUEST_DELAY", 10*time.Millisecond),
		},
		Logging: LoggingConfig{
			LogDir:         getEnv("LOG_DIR", "./logs"),
			AppLogFile:     getEnv("APP_LOG_FILE", "app.log"),
			InvalidLogFile: getEnv("INVALID_LOG_FILE", "invalid_requests.log"),
			DeviceLogDir:   getEnv("DEVICE_LOG_DIR", "./logs/devices"),
		},
	}

	// Validar campos requeridos
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate valida la configuración
func (c *Config) Validate() error {
	if c.MySQL.Host == "" {
		return fmt.Errorf("MYSQL_HOST es requerido")
	}
	if c.MySQL.Database == "" {
		return fmt.Errorf("MYSQL_DATABASE es requerido")
	}
	if c.Redis.Host == "" {
		return fmt.Errorf("REDIS_HOST es requerido")
	}
	if c.MQTT.Enabled && c.MQTT.BrokerURL == "" {
		return fmt.Errorf("MQTT_BROKER_URL es requerido cuando MQTT está habilitado")
	}
	return nil
}

// GetMySQLDSN retorna la cadena de conexión de MySQL
func (c *Config) GetMySQLDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
		c.MySQL.User,
		c.MySQL.Password,
		c.MySQL.Host,
		c.MySQL.Port,
		c.MySQL.Database,
	)
}

// GetRedisAddr retorna la dirección de Redis
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

// Funciones auxiliares para leer variables de entorno con valores por defecto

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getFloat64Env(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
