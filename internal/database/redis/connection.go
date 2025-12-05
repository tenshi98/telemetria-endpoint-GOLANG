package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/config"
)

// Connection representa una conexión a Redis
type Connection struct {
	client *redis.Client
	ttl    time.Duration
}

// NewConnection crea una nueva conexión a Redis
func NewConnection(cfg *config.RedisConfig) (*Connection, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// Probar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("error al hacer ping a Redis: %w", err)
	}

	return &Connection{
		client: client,
		ttl:    cfg.CacheTTL,
	}, nil
}

// GetClient retorna el cliente Redis subyacente
func (c *Connection) GetClient() *redis.Client {
	return c.client
}

// GetTTL retorna el TTL de caché configurado
func (c *Connection) GetTTL() time.Duration {
	return c.ttl
}

// Ping verifica si la conexión a Redis está activa
func (c *Connection) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close cierra la conexión a Redis
func (c *Connection) Close() error {
	return c.client.Close()
}
