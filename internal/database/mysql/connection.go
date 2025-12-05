package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/config"
)

// Connection representa una conexión a la base de datos MySQL
type Connection struct {
	db *sql.DB
}

// NewConnection crea una nueva conexión MySQL
func NewConnection(cfg *config.MySQLConfig) (*Connection, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error al abrir la base de datos: %w", err)
	}

	// Configurar pool de conexiones
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Probar conexión
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error al hacer ping a la base de datos: %w", err)
	}

	return &Connection{db: db}, nil
}

// GetDB retorna la conexión subyacente a la base de datos
func (c *Connection) GetDB() *sql.DB {
	return c.db
}

// Ping verifica si la conexión a la base de datos está activa
func (c *Connection) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// Close cierra la conexión a la base de datos
func (c *Connection) Close() error {
	return c.db.Close()
}
