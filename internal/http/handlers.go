package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/models"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/service"
)

// handleTelemetry maneja los datos de telemetría entrantes vía HTTP POST
func (s *Server) handleTelemetry(c *gin.Context) {
	var req models.TelemetryRequest

	// Vincular solicitud JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		s.logger.Warning("Solicitud JSON inválida: %v", err)

		// Registrar solicitud inválida
		invalidReq := &models.InvalidRequest{
			Timestamp: time.Now(),
			IPAddress: c.ClientIP(),
			Errors: []models.ValidationError{
				{
					Field:   "json",
					Message: "Formato JSON inválido",
				},
			},
		}
		s.logger.LogInvalidRequest(invalidReq)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Formato JSON inválido",
		})
		return
	}

	// Validar campos requeridos
	if err := service.ValidateRequiredFields(&req); err != nil {
		s.logger.Warning("Validación fallida: %v", err)

		// Extraer errores de validación
		var validationErrors []models.ValidationError
		if ve, ok := err.(*service.ValidationErrors); ok {
			validationErrors = ve.GetErrors()
		}

		// Registrar solicitud inválida con detalles
		invalidReq := &models.InvalidRequest{
			Timestamp:     time.Now(),
			IPAddress:     c.ClientIP(),
			Identificador: req.Identificador,
			Latitud:       req.Latitud,
			Longitud:      req.Longitud,
			Errors:        validationErrors,
		}
		s.logger.LogInvalidRequest(invalidReq)

		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Validación fallida",
			"fields": validationErrors,
		})
		return
	}

	// Procesar datos de telemetría
	ctx := c.Request.Context()
	if err := s.telemetryService.ProcessTelemetryData(ctx, &req); err != nil {
		s.logger.Error("Error al procesar datos de telemetría: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error al procesar datos de telemetría",
		})
		return
	}

	// Respuesta exitosa
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Datos de telemetría procesados exitosamente",
	})
}
