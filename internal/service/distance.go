package service

import (
	"math"
)

const (
	// EarthRadiusKm es el radio de la Tierra en kilómetros
	EarthRadiusKm = 6371.0
)

// CalculateDistance calcula la distancia entre dos coordenadas GPS usando la fórmula de Haversine
// Retorna la distancia en metros
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// Convertir grados a radianes
	lat1Rad := degreesToRadians(lat1)
	lon1Rad := degreesToRadians(lon1)
	lat2Rad := degreesToRadians(lat2)
	lon2Rad := degreesToRadians(lon2)

	// Calcular diferencias
	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad

	// Fórmula de Haversine
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distancia en kilómetros
	distanceKm := EarthRadiusKm * c

	// Convertir a metros
	distanceMeters := distanceKm * 1000.0

	return distanceMeters
}

// degreesToRadians convierte grados a radianes
func degreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}
