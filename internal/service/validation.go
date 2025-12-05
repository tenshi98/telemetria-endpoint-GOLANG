package service

import (
	"fmt"

	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/models"
)

// ValidateRequiredFields valida que los campos requeridos estén presentes en la solicitud
func ValidateRequiredFields(req *models.TelemetryRequest) error {
	var errors []models.ValidationError

	if req.Identificador == "" {
		errors = append(errors, models.ValidationError{
			Field:   "identificador",
			Message: "El identificador es requerido",
		})
	}

	if req.Latitud == nil {
		errors = append(errors, models.ValidationError{
			Field:   "latitud",
			Message: "La latitud es requerida",
		})
	}

	if req.Longitud == nil {
		errors = append(errors, models.ValidationError{
			Field:   "longitud",
			Message: "La longitud es requerida",
		})
	}

	if len(errors) > 0 {
		return &ValidationErrors{Errors: errors}
	}

	return nil
}

// ValidationErrors representa múltiples errores de validación
type ValidationErrors struct {
	Errors []models.ValidationError
}

// Error implementa la interfaz error
func (v *ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "Validación fallida"
	}

	msg := "Validación fallida: "
	for i, err := range v.Errors {
		if i > 0 {
			msg += ", "
		}
		msg += fmt.Sprintf("%s: %s", err.Field, err.Message)
	}

	return msg
}

// GetErrors retorna los errores de validación
func (v *ValidationErrors) GetErrors() []models.ValidationError {
	return v.Errors
}
