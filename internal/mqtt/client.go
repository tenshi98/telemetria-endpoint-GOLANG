package mqtt

import (
	"crypto/tls"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/config"
	"github.com/tenshi98/telemetria-endpoint-GOLANG/internal/logger"
)

// Client representa un cliente MQTT
type Client struct {
	client mqtt.Client
	config *config.MQTTConfig
	logger *logger.Logger
}

// NewClient crea un nuevo cliente MQTT
func NewClient(cfg *config.MQTTConfig, log *logger.Logger) (*Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.BrokerURL)
	opts.SetClientID(cfg.ClientID)
	opts.SetCleanSession(cfg.CleanSession)

	// Establecer credenciales si se proporcionan
	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
	}
	if cfg.Password != "" {
		opts.SetPassword(cfg.Password)
	}

	// Configurar TLS si se usa conexión segura
	if cfg.BrokerURL[:3] == "ssl" || cfg.BrokerURL[:3] == "tls" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		opts.SetTLSConfig(tlsConfig)
	}

	// Establecer manejadores de conexión
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Error("Conexión MQTT perdida: %v", err)
	})

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Info("MQTT conectado al broker: %s", cfg.BrokerURL)
	})

	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		log.Warning("MQTT reconectando al broker...")
	})

	// Configuración de reconexión automática
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)

	// Crear cliente
	client := mqtt.NewClient(opts)

	return &Client{
		client: client,
		config: cfg,
		logger: log,
	}, nil
}

// Connect conecta al broker MQTT
func (c *Client) Connect() error {
	c.logger.Info("Conectando al broker MQTT: %s", c.config.BrokerURL)

	token := c.client.Connect()
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("error al conectar al broker MQTT: %w", token.Error())
	}

	c.logger.Info("Conectado exitosamente al broker MQTT")
	return nil
}

// Subscribe se suscribe a un topic con un manejador de mensajes
func (c *Client) Subscribe(handler mqtt.MessageHandler) error {
	c.logger.Info("Suscribiéndose al topic MQTT: %s (QoS: %d)", c.config.Topic, c.config.QoS)

	token := c.client.Subscribe(c.config.Topic, c.config.QoS, handler)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("error al suscribirse al topic: %w", token.Error())
	}

	c.logger.Info("Suscrito exitosamente al topic: %s", c.config.Topic)
	return nil
}

// Disconnect desconecta del broker MQTT
func (c *Client) Disconnect() {
	c.logger.Info("Desconectando del broker MQTT...")
	c.client.Disconnect(250)
	c.logger.Info("Desconectado del broker MQTT")
}

// IsConnected retorna si el cliente está conectado
func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}
