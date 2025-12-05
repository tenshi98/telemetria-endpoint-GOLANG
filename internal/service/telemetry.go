package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/database"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/database/redis"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/logger"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/models"
)

// TelemetryService maneja el procesamiento de datos de telemetría
type TelemetryService struct {
	repo   database.Repository
	cache  database.Cache
	rCache *redis.Cache // Para operaciones adicionales de caché
	logger *logger.Logger
}

// NewTelemetryService crea un nuevo servicio de telemetría
func NewTelemetryService(repo database.Repository, cache database.Cache, rCache *redis.Cache, log *logger.Logger) *TelemetryService {
	return &TelemetryService{
		repo:   repo,
		cache:  cache,
		rCache: rCache,
		logger: log,
	}
}

// ProcessTelemetryData procesa los datos de telemetría entrantes
func (s *TelemetryService) ProcessTelemetryData(ctx context.Context, req *models.TelemetryRequest) error {
	// Validar campos requeridos
	if err := ValidateRequiredFields(req); err != nil {
		return fmt.Errorf("validación fallida: %w", err)
	}

	// Obtener información del dispositivo (caché -> respaldo de base de datos)
	device, err := s.getDevice(ctx, req.Identificador)
	if err != nil {
		return fmt.Errorf("error al obtener dispositivo: %w", err)
	}

	if device == nil {
		// Dispositivo no encontrado - registrar error
		s.logger.Warning("Dispositivo no encontrado: %s", req.Identificador)
		errorRecord := &models.ErrorRecord{
			Identificador: &req.Identificador,
			Fecha:         time.Now(),
			Descripcion:   fmt.Sprintf("Identificador no existe en la base de datos: %s", req.Identificador),
		}
		if err := s.repo.InsertError(ctx, errorRecord); err != nil {
			s.logger.Error("Error al insertar registro de error: %v", err)
		}
		return fmt.Errorf("dispositivo no encontrado: %s", req.Identificador)
	}

	// Validar tiempo fuera de línea
	errors := s.validateOfflineTime(ctx, device)

	// Validar coordenadas faltantes
	coordinateErrors := s.validateCoordinates(req)
	errors = append(errors, coordinateErrors...)

	// Si hay errores, registrarlos
	if len(errors) > 0 {
		errorDesc := strings.Join(errors, "; ")
		errorRecord := &models.ErrorRecord{
			IDTelemetria:  &device.IDTelemetria,
			Identificador: &req.Identificador,
			Fecha:         time.Now(),
			Descripcion:   errorDesc,
		}
		if err := s.repo.InsertError(ctx, errorRecord); err != nil {
			s.logger.Error("Error al insertar registro de error: %v", err)
		}
	}

	// Calcular distancia desde la ubicación anterior
	var distance *float64
	if device.Latitud != nil && device.Longitud != nil && req.Latitud != nil && req.Longitud != nil {
		dist := CalculateDistance(*device.Latitud, *device.Longitud, *req.Latitud, *req.Longitud)
		distance = &dist
	}

	// Crear registro de medición
	measurement := &models.Measurement{
		IDTelemetria: device.IDTelemetria,
		Fecha:        time.Now(),
		Latitud:      req.Latitud,
		Longitud:     req.Longitud,
		Distancia:    distance,
		Sensor1:      req.Sensor1,
		Sensor2:      req.Sensor2,
		Sensor3:      req.Sensor3,
		Sensor4:      req.Sensor4,
		Sensor5:      req.Sensor5,
	}

	// Insertar medición
	if err := s.repo.InsertMeasurement(ctx, measurement); err != nil {
		return fmt.Errorf("error al insertar medición: %w", err)
	}

	// Actualizar tiempo de conexión del dispositivo
	now := time.Now()
	if err := s.repo.UpdateDeviceConnection(ctx, device.IDTelemetria, now); err != nil {
		s.logger.Warning("Error al actualizar conexión del dispositivo: %v", err)
	}

	// Actualizar caché con nueva ubicación y marca de tiempo
	if req.Latitud != nil && req.Longitud != nil {
		if err := s.rCache.UpdateDeviceLocation(ctx, req.Identificador, *req.Latitud, *req.Longitud, now); err != nil {
			s.logger.Warning("Error al actualizar ubicación del dispositivo en caché: %v", err)
		}
	}

	// Registrar datos del dispositivo
	if err := s.logger.LogDeviceData(req.Identificador, req); err != nil {
		s.logger.Warning("Error al registrar datos del dispositivo: %v", err)
	}

	s.logger.Info("Datos de telemetría procesados para dispositivo: %s", req.Identificador)
	return nil
}

// getDevice obtiene información del dispositivo desde caché o base de datos
func (s *TelemetryService) getDevice(ctx context.Context, identifier string) (*models.Device, error) {
	// Intentar caché primero
	device, err := s.cache.GetDevice(ctx, identifier)
	if err != nil {
		s.logger.Warning("Error al obtener dispositivo del caché: %v", err)
	}

	if device != nil {
		s.logger.Info("Dispositivo encontrado en caché: %s", identifier)
		return device, nil
	}

	// Respaldo a base de datos
	s.logger.Info("Dispositivo no está en caché, consultando base de datos: %s", identifier)
	device, err = s.repo.GetDeviceByIdentifier(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("error al consultar dispositivo desde la base de datos: %w", err)
	}

	if device == nil {
		return nil, nil // Dispositivo no encontrado
	}

	// Poblar caché
	if err := s.cache.SetDevice(ctx, device); err != nil {
		s.logger.Warning("Error al almacenar dispositivo en caché: %v", err)
	} else {
		s.logger.Info("Dispositivo almacenado en caché: %s", identifier)
	}

	return device, nil
}

// validateOfflineTime verifica si el dispositivo ha estado fuera de línea más tiempo del permitido
func (s *TelemetryService) validateOfflineTime(ctx context.Context, device *models.Device) []string {
	var errors []string

	// Parsear TiempoFueraLinea (formato: HH:MM:SS)
	maxOfflineDuration, err := parseTimeDuration(device.TiempoFueraLinea)
	if err != nil {
		s.logger.Warning("Error al parsear TiempoFueraLinea para dispositivo %s: %v", device.Identificador, err)
		return errors
	}

	// Calcular tiempo desde la última conexión
	timeSinceLastConnection := time.Since(device.UltimaConexion)

	// Verificar si se excedió el tiempo fuera de línea
	if timeSinceLastConnection > maxOfflineDuration {
		errorMsg := fmt.Sprintf("Tiempo fuera de línea excedido: última conexión %s, tiempo máximo permitido %s, tiempo transcurrido %s",
			device.UltimaConexion.Format(time.RFC3339),
			device.TiempoFueraLinea,
			timeSinceLastConnection.String(),
		)
		errors = append(errors, errorMsg)
		s.logger.Warning("Dispositivo %s tiempo fuera de línea excedido: %s", device.Identificador, errorMsg)
	}

	return errors
}

// validateCoordinates verifica si faltan coordenadas
func (s *TelemetryService) validateCoordinates(req *models.TelemetryRequest) []string {
	var errors []string

	if req.Latitud == nil {
		errors = append(errors, "Falta dato: Latitud")
	}

	if req.Longitud == nil {
		errors = append(errors, "Falta dato: Longitud")
	}

	return errors
}

// parseTimeDuration parsea una cadena de tiempo en formato HH:MM:SS a time.Duration
func parseTimeDuration(timeStr string) (time.Duration, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("formato de tiempo inválido: %s", timeStr)
	}

	var hours, minutes, seconds int
	_, err := fmt.Sscanf(timeStr, "%d:%d:%d", &hours, &minutes, &seconds)
	if err != nil {
		return 0, fmt.Errorf("error al parsear tiempo: %w", err)
	}

	duration := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second

	return duration, nil
}
