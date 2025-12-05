package database

import (
	"context"
	"time"

	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/models"
)

// Repository define la interfaz para operaciones de base de datos
// Esta abstracción permite una fácil migración a otras bases de datos SQL
type Repository interface {
	// Operaciones de dispositivos
	GetDeviceByIdentifier(ctx context.Context, identifier string) (*models.Device, error)
	UpdateDeviceConnection(ctx context.Context, deviceID uint, timestamp time.Time) error

	// Operaciones de mediciones
	InsertMeasurement(ctx context.Context, measurement *models.Measurement) error

	// Operaciones de errores
	InsertError(ctx context.Context, errorRecord *models.ErrorRecord) error

	// Verificación de salud
	Ping(ctx context.Context) error

	// Cerrar conexión
	Close() error
}

// Cache define la interfaz para operaciones de caché
type Cache interface {
	// Operaciones de caché de dispositivos
	GetDevice(ctx context.Context, identifier string) (*models.Device, error)
	SetDevice(ctx context.Context, device *models.Device) error
	DeleteDevice(ctx context.Context, identifier string) error

	// Verificación de salud
	Ping(ctx context.Context) error

	// Cerrar conexión
	Close() error
}
