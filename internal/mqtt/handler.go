package mqtt

import (
	"context"
	"encoding/json"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/logger"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/models"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/service"
)

// Handler maneja mensajes MQTT
type Handler struct {
	telemetryService *service.TelemetryService
	logger           *logger.Logger
}

// NewHandler crea un nuevo manejador de mensajes MQTT
func NewHandler(telemetryService *service.TelemetryService, log *logger.Logger) *Handler {
	return &Handler{
		telemetryService: telemetryService,
		logger:           log,
	}
}

// HandleMessage maneja mensajes MQTT entrantes
func (h *Handler) HandleMessage(client mqtt.Client, msg mqtt.Message) {
	h.logger.Info("Mensaje MQTT recibido en topic: %s", msg.Topic())

	// Parsear payload del mensaje
	var req models.TelemetryRequest
	if err := json.Unmarshal(msg.Payload(), &req); err != nil {
		h.logger.Error("Error al parsear mensaje MQTT: %v", err)

		// Registrar solicitud inválida
		invalidReq := &models.InvalidRequest{
			Timestamp: time.Now(),
			IPAddress: "MQTT",
			Errors: []models.ValidationError{
				{
					Field:   "json",
					Message: "Formato JSON inválido",
				},
			},
		}
		h.logger.LogInvalidRequest(invalidReq)
		return
	}

	// Validar campos requeridos
	if err := service.ValidateRequiredFields(&req); err != nil {
		h.logger.Warning("Validación de mensaje MQTT fallida: %v", err)

		// Extraer errores de validación
		var validationErrors []models.ValidationError
		if ve, ok := err.(*service.ValidationErrors); ok {
			validationErrors = ve.GetErrors()
		}

		// Registrar solicitud inválida
		invalidReq := &models.InvalidRequest{
			Timestamp:     time.Now(),
			IPAddress:     "MQTT",
			Identificador: req.Identificador,
			Latitud:       req.Latitud,
			Longitud:      req.Longitud,
			Errors:        validationErrors,
		}
		h.logger.LogInvalidRequest(invalidReq)
		return
	}

	// Procesar datos de telemetría
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := h.telemetryService.ProcessTelemetryData(ctx, &req); err != nil {
		h.logger.Error("Error al procesar datos de telemetría MQTT: %v", err)
		return
	}

	h.logger.Info("Datos de telemetría MQTT procesados exitosamente para dispositivo: %s", req.Identificador)
}
