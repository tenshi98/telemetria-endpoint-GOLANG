package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/models"
)

// Cache implementa la interfaz database.Cache para Redis
type Cache struct {
	conn *Connection
}

// NewCache crea un nuevo caché Redis
func NewCache(conn *Connection) *Cache {
	return &Cache{conn: conn}
}

// GetDevice obtiene un dispositivo del caché por identificador
func (c *Cache) GetDevice(ctx context.Context, identifier string) (*models.Device, error) {
	key := fmt.Sprintf("device:%s", identifier)

	// Obtener todos los campos del hash
	result, err := c.conn.GetClient().HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("error al obtener dispositivo del caché: %w", err)
	}

	if len(result) == 0 {
		return nil, nil // Dispositivo no está en caché
	}

	// Parsear datos del dispositivo
	device := &models.Device{
		Identificador: identifier,
	}

	// Parsear idTelemetria
	if val, ok := result["idTelemetria"]; ok {
		var id uint
		if err := json.Unmarshal([]byte(val), &id); err == nil {
			device.IDTelemetria = id
		}
	}

	// Parsear nombre
	if val, ok := result["nombre"]; ok {
		device.Nombre = val
	}

	// Parsear ultimaConexion
	if val, ok := result["ultimaConexion"]; ok {
		if t, err := time.Parse(time.RFC3339, val); err == nil {
			device.UltimaConexion = t
		}
	}

	// Parsear tiempoFueraLinea
	if val, ok := result["tiempoFueraLinea"]; ok {
		device.TiempoFueraLinea = val
	}

	// Parsear latitud
	if val, ok := result["latitud"]; ok {
		var lat float64
		if err := json.Unmarshal([]byte(val), &lat); err == nil {
			device.Latitud = &lat
		}
	}

	// Parsear longitud
	if val, ok := result["longitud"]; ok {
		var lon float64
		if err := json.Unmarshal([]byte(val), &lon); err == nil {
			device.Longitud = &lon
		}
	}

	return device, nil
}

// SetDevice almacena un dispositivo en caché
func (c *Cache) SetDevice(ctx context.Context, device *models.Device) error {
	key := fmt.Sprintf("device:%s", device.Identificador)

	// Preparar datos para el hash
	data := map[string]interface{}{
		"idTelemetria":     device.IDTelemetria,
		"nombre":           device.Nombre,
		"ultimaConexion":   device.UltimaConexion.Format(time.RFC3339),
		"tiempoFueraLinea": device.TiempoFueraLinea,
	}

	if device.Latitud != nil {
		data["latitud"] = *device.Latitud
	}

	if device.Longitud != nil {
		data["longitud"] = *device.Longitud
	}

	// Almacenar en hash
	if err := c.conn.GetClient().HSet(ctx, key, data).Err(); err != nil {
		return fmt.Errorf("error al establecer dispositivo en caché: %w", err)
	}

	// Establecer expiración
	if err := c.conn.GetClient().Expire(ctx, key, c.conn.GetTTL()).Err(); err != nil {
		return fmt.Errorf("error al establecer expiración del caché: %w", err)
	}

	return nil
}

// DeleteDevice elimina un dispositivo del caché
func (c *Cache) DeleteDevice(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("device:%s", identifier)

	if err := c.conn.GetClient().Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("error al eliminar dispositivo del caché: %w", err)
	}

	return nil
}

// UpdateDeviceLocation actualiza solo los campos de ubicación en caché
func (c *Cache) UpdateDeviceLocation(ctx context.Context, identifier string, latitud, longitud float64, timestamp time.Time) error {
	key := fmt.Sprintf("device:%s", identifier)

	data := map[string]interface{}{
		"latitud":        latitud,
		"longitud":       longitud,
		"ultimaConexion": timestamp.Format(time.RFC3339),
	}

	if err := c.conn.GetClient().HSet(ctx, key, data).Err(); err != nil {
		return fmt.Errorf("error al actualizar ubicación del dispositivo en caché: %w", err)
	}

	// Refrescar expiración
	if err := c.conn.GetClient().Expire(ctx, key, c.conn.GetTTL()).Err(); err != nil {
		return fmt.Errorf("error al refrescar expiración del caché: %w", err)
	}

	return nil
}

// Ping verifica si la conexión a Redis está activa
func (c *Cache) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

// Close cierra la conexión a Redis
func (c *Cache) Close() error {
	return c.conn.Close()
}
