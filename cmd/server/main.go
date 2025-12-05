package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/config"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/database/mysql"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/database/redis"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/http"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/logger"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/mqtt"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/service"
)

func main() {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al cargar la configuración: %v\n", err)
		os.Exit(1)
	}

	// Inicializar logger
	log, err := logger.New(
		cfg.Logging.LogDir,
		cfg.Logging.AppLogFile,
		cfg.Logging.InvalidLogFile,
		cfg.Logging.DeviceLogDir,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al inicializar el logger: %v\n", err)
		os.Exit(1)
	}

	log.Info("Iniciando Servidor de Endpoint de Telemetría...")
	log.Info("Configuración cargada exitosamente")

	// Inicializar conexión MySQL
	log.Info("Conectando a la base de datos MySQL...")
	mysqlConn, err := mysql.NewConnection(&cfg.MySQL)
	if err != nil {
		log.Error("Error al conectar a MySQL: %v", err)
		os.Exit(1)
	}
	defer mysqlConn.Close()
	log.Info("Conexión MySQL establecida")

	// Inicializar repositorio MySQL
	repo := mysql.NewRepository(mysqlConn)

	// Inicializar conexión Redis
	log.Info("Conectando a Redis...")
	redisConn, err := redis.NewConnection(&cfg.Redis)
	if err != nil {
		log.Error("Error al conectar a Redis: %v", err)
		os.Exit(1)
	}
	defer redisConn.Close()
	log.Info("Conexión Redis establecida")

	// Inicializar caché Redis
	cache := redis.NewCache(redisConn)

	// Inicializar servicio de telemetría
	telemetryService := service.NewTelemetryService(repo, cache, cache, log)
	log.Info("Servicio de telemetría inicializado")

	// Inicializar servidor HTTP
	httpServer := http.NewServer(&cfg.Server, &cfg.RateLimit, telemetryService, log)

	// Iniciar servidor HTTP en una goroutine
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Error("Error del servidor HTTP: %v", err)
			os.Exit(1)
		}
	}()

	// Inicializar e iniciar cliente MQTT si está habilitado
	var mqttClient *mqtt.Client
	if cfg.MQTT.Enabled {
		log.Info("MQTT está habilitado, inicializando cliente MQTT...")
		mqttClient, err = mqtt.NewClient(&cfg.MQTT, log)
		if err != nil {
			log.Error("Error al crear cliente MQTT: %v", err)
			os.Exit(1)
		}

		// Conectar al broker MQTT
		if err := mqttClient.Connect(); err != nil {
			log.Error("Error al conectar al broker MQTT: %v", err)
			os.Exit(1)
		}

		// Crear manejador MQTT
		mqttHandler := mqtt.NewHandler(telemetryService, log)

		// Suscribirse al topic
		if err := mqttClient.Subscribe(mqttHandler.HandleMessage); err != nil {
			log.Error("Error al suscribirse al topic MQTT: %v", err)
			mqttClient.Disconnect()
			os.Exit(1)
		}

		log.Info("Cliente MQTT iniciado y suscrito al topic")
	} else {
		log.Info("MQTT está deshabilitado")
	}

	// Esperar señal de interrupción para apagar ordenadamente
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Señal de apagado recibida, apagando ordenadamente...")

	// Apagar servidor HTTP
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("Error al apagar servidor HTTP: %v", err)
	}

	// Desconectar cliente MQTT si está habilitado
	if mqttClient != nil {
		mqttClient.Disconnect()
	}

	log.Info("Apagado del servidor completado")
}
