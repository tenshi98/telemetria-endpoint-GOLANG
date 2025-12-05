package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/models"
)

// Logger maneja todas las operaciones de logging para la aplicación
type Logger struct {
	appLogger     *log.Logger
	invalidLogger *log.Logger
	deviceLoggers map[string]*log.Logger
	logDir        string
	deviceLogDir  string
	mu            sync.RWMutex
}

// New crea una nueva instancia de Logger
func New(logDir, appLogFile, invalidLogFile, deviceLogDir string) (*Logger, error) {
	// Crear directorios de logs
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("error al crear directorio de logs: %w", err)
	}

	if err := os.MkdirAll(deviceLogDir, 0755); err != nil {
		return nil, fmt.Errorf("error al crear directorio de logs de dispositivos: %w", err)
	}

	// Crear archivo de log de aplicación
	appLogPath := filepath.Join(logDir, appLogFile)
	appFile, err := os.OpenFile(appLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("error al abrir archivo de log de aplicación: %w", err)
	}

	// Crear archivo de log de solicitudes inválidas
	invalidLogPath := filepath.Join(logDir, invalidLogFile)
	invalidFile, err := os.OpenFile(invalidLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("error al abrir archivo de log de solicitudes inválidas: %w", err)
	}

	return &Logger{
		appLogger:     log.New(appFile, "", log.LstdFlags),
		invalidLogger: log.New(invalidFile, "", log.LstdFlags),
		deviceLoggers: make(map[string]*log.Logger),
		logDir:        logDir,
		deviceLogDir:  deviceLogDir,
	}, nil
}

// Info registra un mensaje informativo
func (l *Logger) Info(format string, v ...interface{}) {
	l.appLogger.Printf("[INFO] "+format, v...)
}

// Warning registra un mensaje de advertencia
func (l *Logger) Warning(format string, v ...interface{}) {
	l.appLogger.Printf("[WARNING] "+format, v...)
}

// Error registra un mensaje de error
func (l *Logger) Error(format string, v ...interface{}) {
	l.appLogger.Printf("[ERROR] "+format, v...)
}

// LogDeviceData registra datos de telemetría para un dispositivo específico
func (l *Logger) LogDeviceData(identifier string, data *models.TelemetryRequest) error {
	deviceLogger, err := l.getDeviceLogger(identifier)
	if err != nil {
		return fmt.Errorf("error al obtener logger de dispositivo: %w", err)
	}

	// Formatear entrada de log
	logEntry := fmt.Sprintf("Identificador: %s, Latitud: %v, Longitud: %v",
		identifier,
		formatFloat(data.Latitud),
		formatFloat(data.Longitud),
	)

	// Agregar datos opcionales de sensores
	if data.Sensor1 != nil {
		logEntry += fmt.Sprintf(", Sensor_1: %.6f", *data.Sensor1)
	}
	if data.Sensor2 != nil {
		logEntry += fmt.Sprintf(", Sensor_2: %.6f", *data.Sensor2)
	}
	if data.Sensor3 != nil {
		logEntry += fmt.Sprintf(", Sensor_3: %.6f", *data.Sensor3)
	}
	if data.Sensor4 != nil {
		logEntry += fmt.Sprintf(", Sensor_4: %.6f", *data.Sensor4)
	}
	if data.Sensor5 != nil {
		logEntry += fmt.Sprintf(", Sensor_5: %.6f", *data.Sensor5)
	}

	deviceLogger.Println(logEntry)
	return nil
}

// LogInvalidRequest registra una solicitud con campos requeridos faltantes
func (l *Logger) LogInvalidRequest(req *models.InvalidRequest) {
	logEntry := fmt.Sprintf("IP: %s, Timestamp: %s",
		req.IPAddress,
		req.Timestamp.Format(time.RFC3339),
	)

	// Agregar identificador si está presente
	if req.Identificador != "" {
		logEntry += fmt.Sprintf(", Identificador: %s", req.Identificador)
	} else {
		logEntry += ", Identificador: MISSING"
	}

	// Agregar coordenadas si están presentes
	if req.Latitud != nil {
		logEntry += fmt.Sprintf(", Latitud: %.6f", *req.Latitud)
	} else {
		logEntry += ", Latitud: MISSING"
	}

	if req.Longitud != nil {
		logEntry += fmt.Sprintf(", Longitud: %.6f", *req.Longitud)
	} else {
		logEntry += ", Longitud: MISSING"
	}

	// Agregar errores de validación
	if len(req.Errors) > 0 {
		logEntry += ", Errors: ["
		for i, err := range req.Errors {
			if i > 0 {
				logEntry += ", "
			}
			logEntry += fmt.Sprintf("%s: %s", err.Field, err.Message)
		}
		logEntry += "]"
	}

	l.invalidLogger.Println(logEntry)
}

// getDeviceLogger obtiene o crea un logger para un dispositivo específico
func (l *Logger) getDeviceLogger(identifier string) (*log.Logger, error) {
	l.mu.RLock()
	if logger, exists := l.deviceLoggers[identifier]; exists {
		l.mu.RUnlock()
		return logger, nil
	}
	l.mu.RUnlock()

	// Crear nuevo logger de dispositivo
	l.mu.Lock()
	defer l.mu.Unlock()

	// Verificar nuevamente después de adquirir el bloqueo de escritura
	if logger, exists := l.deviceLoggers[identifier]; exists {
		return logger, nil
	}

	// Crear archivo de log de dispositivo
	deviceLogPath := filepath.Join(l.deviceLogDir, fmt.Sprintf("%s.log", identifier))
	deviceFile, err := os.OpenFile(deviceLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("error al abrir archivo de log de dispositivo: %w", err)
	}

	deviceLogger := log.New(deviceFile, "", log.LstdFlags)
	l.deviceLoggers[identifier] = deviceLogger

	return deviceLogger, nil
}

// formatFloat formatea un puntero float para logging
func formatFloat(f *float64) string {
	if f == nil {
		return "null"
	}
	return fmt.Sprintf("%.6f", *f)
}
