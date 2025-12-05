package http

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/config"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/logger"
	"golang.org/x/time/rate"
)

// LoggingMiddleware registra solicitudes HTTP
func LoggingMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Procesar solicitud
		c.Next()

		// Registrar detalles de la solicitud
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		log.Info("%s %s - Status: %d - Duration: %v - IP: %s",
			method,
			path,
			statusCode,
			duration,
			c.ClientIP(),
		)
	}
}

// RateLimitMiddleware implementa limitación de tasa usando algoritmo de cubo de tokens
func RateLimitMiddleware(cfg *config.RateLimitConfig) gin.HandlerFunc {
	// Crear un limitador de tasa por dirección IP
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Limpiar clientes antiguos periódicamente
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		if _, exists := clients[ip]; !exists {
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.BurstSize),
			}
		}
		clients[ip].lastSeen = time.Now()
		limiter := clients[ip].limiter
		mu.Unlock()

		// Verificar si la solicitud está permitida
		if !limiter.Allow() {
			c.JSON(429, gin.H{
				"error": "Límite de tasa excedido",
			})
			c.Abort()
			return
		}

		// Agregar retraso de solicitud si está configurado
		if cfg.RequestDelay > 0 {
			time.Sleep(cfg.RequestDelay)
		}

		c.Next()
	}
}
