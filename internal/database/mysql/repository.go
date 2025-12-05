package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/models"
)

// Repository implementa la interfaz database.Repository para MySQL
type Repository struct {
	conn *Connection
}

// NewRepository crea un nuevo repositorio MySQL
func NewRepository(conn *Connection) *Repository {
	return &Repository{conn: conn}
}

// GetDeviceByIdentifier obtiene un dispositivo por su identificador
func (r *Repository) GetDeviceByIdentifier(ctx context.Context, identifier string) (*models.Device, error) {
	query := `
		SELECT idTelemetria, Identificador, Nombre, UltimaConexion, TiempoFueraLinea
		FROM equipos_telemetria
		WHERE Identificador = ?
	`

	var device models.Device
	err := r.conn.GetDB().QueryRowContext(ctx, query, identifier).Scan(
		&device.IDTelemetria,
		&device.Identificador,
		&device.Nombre,
		&device.UltimaConexion,
		&device.TiempoFueraLinea,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Dispositivo no encontrado
	}
	if err != nil {
		return nil, fmt.Errorf("error al consultar dispositivo: %w", err)
	}

	return &device, nil
}

// UpdateDeviceConnection actualiza la marca de tiempo de la última conexión de un dispositivo
func (r *Repository) UpdateDeviceConnection(ctx context.Context, deviceID uint, timestamp time.Time) error {
	query := `
		UPDATE equipos_telemetria
		SET UltimaConexion = ?
		WHERE idTelemetria = ?
	`

	result, err := r.conn.GetDB().ExecContext(ctx, query, timestamp, deviceID)
	if err != nil {
		return fmt.Errorf("error al actualizar conexión del dispositivo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error al obtener filas afectadas: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dispositivo no encontrado: %d", deviceID)
	}

	return nil
}

// InsertMeasurement inserta un nuevo registro de medición
func (r *Repository) InsertMeasurement(ctx context.Context, measurement *models.Measurement) error {
	query := `
		INSERT INTO equipos_telemetria_datos 
		(idTelemetria, Fecha, Latitud, Longitud, Distancia, Sensor_1, Sensor_2, Sensor_3, Sensor_4, Sensor_5)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.conn.GetDB().ExecContext(ctx, query,
		measurement.IDTelemetria,
		measurement.Fecha,
		measurement.Latitud,
		measurement.Longitud,
		measurement.Distancia,
		measurement.Sensor1,
		measurement.Sensor2,
		measurement.Sensor3,
		measurement.Sensor4,
		measurement.Sensor5,
	)
	if err != nil {
		return fmt.Errorf("error al insertar medición: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error al obtener último ID insertado: %w", err)
	}

	measurement.IDMedicion = uint64(id)
	return nil
}

// InsertError inserta un nuevo registro de error
func (r *Repository) InsertError(ctx context.Context, errorRecord *models.ErrorRecord) error {
	query := `
		INSERT INTO equipos_telemetria_errores 
		(idTelemetria, Identificador, Fecha, descripcion)
		VALUES (?, ?, ?, ?)
	`

	result, err := r.conn.GetDB().ExecContext(ctx, query,
		errorRecord.IDTelemetria,
		errorRecord.Identificador,
		errorRecord.Fecha,
		errorRecord.Descripcion,
	)
	if err != nil {
		return fmt.Errorf("error al insertar error: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error al obtener último ID insertado: %w", err)
	}

	errorRecord.IDMedicion = uint64(id)
	return nil
}

// Ping verifica si la conexión a la base de datos está activa
func (r *Repository) Ping(ctx context.Context) error {
	return r.conn.Ping(ctx)
}

// Close cierra la conexión a la base de datos
func (r *Repository) Close() error {
	return r.conn.Close()
}
