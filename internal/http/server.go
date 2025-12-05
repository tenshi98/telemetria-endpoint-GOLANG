package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/config"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/logger"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/service"
)

// Server representa el servidor HTTP
type Server struct {
	router           *gin.Engine
	server           *http.Server
	telemetryService *service.TelemetryService
	logger           *logger.Logger
	config           *config.ServerConfig
}

// NewServer crea un nuevo servidor HTTP
func NewServer(cfg *config.ServerConfig, rateLimitCfg *config.RateLimitConfig, telemetryService *service.TelemetryService, log *logger.Logger) *Server {
	// Establecer modo Gin a release para producción
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Agregar middleware
	router.Use(gin.Recovery())
	router.Use(LoggingMiddleware(log))
	router.Use(RateLimitMiddleware(rateLimitCfg))

	s := &Server{
		router:           router,
		telemetryService: telemetryService,
		logger:           log,
		config:           cfg,
	}

	// Registrar rutas
	s.registerRoutes()

	return s
}

// registerRoutes registra todas las rutas HTTP
func (s *Server) registerRoutes() {
	// Endpoint de verificación de salud
	s.router.GET("/health", s.healthCheck)

	// Endpoint de datos de telemetría
	s.router.POST("/telemetry", s.handleTelemetry)
}

// Start inicia el servidor HTTP
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	s.logger.Info("Iniciando servidor HTTP en puerto %s", s.config.Port)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("error al iniciar servidor: %w", err)
	}

	return nil
}

// Shutdown apaga el servidor HTTP de forma ordenada
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Apagando servidor HTTP...")
	return s.server.Shutdown(ctx)
}

// healthCheck maneja solicitudes de verificación de salud
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
