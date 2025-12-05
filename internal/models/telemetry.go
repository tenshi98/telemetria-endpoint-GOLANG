package models

import (
	"time"
)

// TelemetryRequest representa los datos de telemetría entrantes
type TelemetryRequest struct {
	Identificador string   `json:"identificador" binding:"required"`
	Latitud       *float64 `json:"latitud" binding:"required"`
	Longitud      *float64 `json:"longitud" binding:"required"`
	Sensor1       *float64 `json:"sensor_1"`
	Sensor2       *float64 `json:"sensor_2"`
	Sensor3       *float64 `json:"sensor_3"`
	Sensor4       *float64 `json:"sensor_4"`
	Sensor5       *float64 `json:"sensor_5"`
}

// Device representa un dispositivo de telemetría
type Device struct {
	IDTelemetria      uint      `json:"idTelemetria"`
	Identificador     string    `json:"identificador"`
	Nombre            string    `json:"nombre"`
	UltimaConexion    time.Time `json:"ultimaConexion"`
	TiempoFueraLinea  string    `json:"tiempoFueraLinea"` // Formato TIME HH:MM:SS
	Latitud           *float64  `json:"latitud"`
	Longitud          *float64  `json:"longitud"`
}

// Measurement representa una medición de telemetría
type Measurement struct {
	IDMedicion   uint64    `json:"idMedicion"`
	IDTelemetria uint      `json:"idTelemetria"`
	Fecha        time.Time `json:"fecha"`
	Latitud      *float64  `json:"latitud"`
	Longitud     *float64  `json:"longitud"`
	Distancia    *float64  `json:"distancia"`
	Sensor1      *float64  `json:"sensor_1"`
	Sensor2      *float64  `json:"sensor_2"`
	Sensor3      *float64  `json:"sensor_3"`
	Sensor4      *float64  `json:"sensor_4"`
	Sensor5      *float64  `json:"sensor_5"`
}

// ErrorRecord representa una entrada de error
type ErrorRecord struct {
	IDMedicion    uint64    `json:"idMedicion"`
	IDTelemetria  *uint     `json:"idTelemetria"`
	Identificador *string   `json:"identificador"`
	Fecha         time.Time `json:"fecha"`
	Descripcion   string    `json:"descripcion"`
}

// ValidationError representa errores de validación
type ValidationError struct {
	Field   string
	Message string
}

// InvalidRequest representa una solicitud con campos requeridos faltantes
type InvalidRequest struct {
	Timestamp     time.Time
	IPAddress     string
	Identificador string
	Latitud       *float64
	Longitud      *float64
	Errors        []ValidationError
}
