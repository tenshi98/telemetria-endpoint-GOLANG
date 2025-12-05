# Telemetry Endpoint - Golang

Sistema completo de recepción y procesamiento de datos de telemetría desarrollado en Go 1.22+, con soporte para HTTP POST y MQTT, almacenamiento en MySQL, caché en Redis, y logging exhaustivo.

## 📋 Tabla de Contenidos

- [Características](#-características)
- [Requisitos](#-requisitos)
- [Instalación](#-instalación)
- [Configuración](#-configuración)
- [Ejecución](#-ejecución)
- [Uso](#-uso)
- [Estructura del Proyecto](#-estructura-del-proyecto)
- [Módulos](#-módulos)
- [Migración a Otras Bases de Datos](#-migración-a-otras-bases-de-datos)
- [Buenas Prácticas Implementadas](#-buenas-prácticas-implementadas)
- [Logs](#-logs)
- [Solución de Problemas](#-solución-de-problemas)

## ✨ Características

- ✅ **Recepción de datos**: Soporta HTTP POST y MQTT
- ✅ **Caché Redis**: Almacenamiento en caché de dispositivos para consultas rápidas
- ✅ **Rate Limiting**: Control de límite de peticiones por dispositivo configurable
- ✅ **Validación Robusta**: Validación completa de datos de entrada
- ✅ **Cálculo de Distancia**: Fórmula de Haversine para cálculo preciso de distancias
- ✅ **Logging Completo**: Logs por dispositivo, requests inválidos, sistema y errores
- ✅ **Arquitectura Modular**: Fácil mantenimiento y extensión
- ✅ **Abstracción de base de datos**: Migración simple a otros motores de base de datos
- ✅ **Manejo de Errores**: Registro de errores en base de datos y archivos
- ✅ **Validación de tiempo offline**: Detección de dispositivos fuera de línea
- ✅ **Graceful shutdown**: Cierre ordenado de conexiones

## 🛠️ Requisitos

### Software Requerido
- **Go**: 1.22 o superior
- **MySQL**: 5.7 o superior (o MariaDB 10.2+)
- **Redis**: 6.0 o superior
- **MQTT Broker** (opcional): Mosquitto 1.4+ (u otro broker compatible)

## 📦 Instalación

### 1. Clonar o descargar el proyecto

```bash
git clone https://github.com/tenshi98/telemetria-endpoint-GOLANG.git
cd telemetria-endpoint-GOLANG
```

### 2. Instalación de Go

Si Go no está instalado en tu sistema:

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# Verificar instalación
go version
```

Para versiones más recientes de Go, descarga desde [golang.org](https://golang.org/dl/).

### 3. Inicializar módulo Go

```bash
go mod init github.com/tenshi98/telemetria-endpoint-GOLANG
```

### 4. Instalar dependencias

```bash
go mod tidy
```

### 5. Instalar Base de Datos

```bash
# Conectar a MySQL
mysql -u root -p

# Ejecutar schema
mysql -u root -p < migrations/schema.sql

# (Opcional) Cargar datos de prueba
mysql -u root -p < migrations/seed.sql
```

### 6. Instalar Redis (opcional)

```bash
# Ubuntu/Debian
sudo apt install redis-server

# Iniciar Redis
sudo systemctl start redis-server
sudo systemctl enable redis-server
# o
redis-server

# Verificar que Redis esté corriendo
redis-cli ping
# Debe responder: PONG
```

### 7. Instalar Mosquitto (Broker MQTT - opcional)

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y mosquitto mosquitto-clients

# Iniciar Mosquitto
sudo systemctl start mosquitto
sudo systemctl enable mosquitto

# Verificar que está corriendo
sudo systemctl status mosquitto
mosquitto_pub -h localhost -t test -m "hello"

# Probar
mosquitto_sub -t "telemetry/data" -v
```

## ⚙️ Configuración

### 1. Configurar variables de entorno

```bash
# Copiar archivo de ejemplo
cp .env.example .env

# Editar .env con tus configuraciones
nano .env
```

**Configuraciones importantes en `.env`:**

```env
# MySQL
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_DATABASE=telemetria
MYSQL_USER=root
MYSQL_PASSWORD=tu_contraseña

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# MQTT (opcional)
MQTT_ENABLED=false
```

**Configuración completa con MQTT:**

```env
# Server
SERVER_PORT=8080

# MySQL
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=tu_contraseña
MYSQL_DATABASE=telemetria

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_CACHE_TTL=24h

# MQTT
MQTT_ENABLED=true
MQTT_BROKER_URL=tcp://localhost:1883
MQTT_CLIENT_ID=telemetry-endpoint
MQTT_TOPIC=telemetry/data
MQTT_QOS=1

# Rate Limiting
RATE_LIMIT_RPS=100.0
RATE_LIMIT_BURST=200
REQUEST_DELAY=10ms

# Logging
LOG_DIR=./logs
DEVICE_LOG_DIR=./logs/devices
```

### 2. Cargar variables de entorno

```bash
# Opción 1: Exportar manualmente
export $(cat .env | xargs)

# Opción 2: Usar un cargador de .env (recomendado para producción)
# Instalar: go get github.com/joho/godotenv
```

## 🏃 Ejecución

### Modo Desarrollo

```bash
# Compilar y ejecutar
go run cmd/server/main.go
```

### Modo Producción

```bash
# Compilar binario
go build -o telemetry-server cmd/server/main.go

# Ejecutar
./telemetry-server
```

### Con systemd (Linux)

Crear archivo `/etc/systemd/system/telemetry.service`:

```ini
[Unit]
Description=Telemetry Endpoint Server
After=network.target mysql.service redis.service

[Service]
Type=simple
User=tu_usuario
WorkingDirectory=/ruta/al/proyecto
EnvironmentFile=/ruta/al/proyecto/.env
ExecStart=/ruta/al/proyecto/telemetry-server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl start telemetry
sudo systemctl enable telemetry
sudo systemctl status telemetry
```

## 📡 Uso

### HTTP POST

**Endpoint:** `POST http://localhost:8080/telemetry`

**Headers:**
```
Content-Type: application/json
```

**Body (campos requeridos):**
```json
{
  "identificador": "DEVICE001",
  "latitud": -34.603722,
  "longitud": -58.381592
}
```

**Body (completo con sensores):**
```json
{
  "identificador": "DEVICE001",
  "latitud": -34.603722,
  "longitud": -58.381592,
  "sensor_1": 23.5,
  "sensor_2": 45.2,
  "sensor_3": 67.8,
  "sensor_4": 12.3,
  "sensor_5": 89.1
}
```

**Ejemplo con curl:**
```bash
curl -X POST http://localhost:8080/telemetry \
  -H "Content-Type: application/json" \
  -d '{
    "identificador": "DEVICE001",
    "latitud": -34.603722,
    "longitud": -58.381592,
    "sensor_1": 23.5
  }'
```

**Respuesta exitosa:**
```json
{
  "status": "success",
  "message": "Telemetry data processed successfully"
}
```

**Respuesta de error (validación):**
```json
{
  "error": "Validación fallida",
  "fields": [
    {
      "Field": "latitud",
      "Message": "La latitud es requerida"
    }
  ]
}
```

### MQTT

**Publicar datos:**
```bash
mosquitto_pub -t "telemetry/data" -m '{
  "identificador": "DEVICE001",
  "latitud": -34.603722,
  "longitud": -58.381592,
  "sensor_1": 23.5
}'
```

**Ejemplo con Python (paho-mqtt):**
```python
import paho.mqtt.client as mqtt
import json

client = mqtt.Client()
client.connect("localhost", 1883, 60)

data = {
    "identificador": "DEVICE001",
    "latitud": -34.603722,
    "longitud": -58.381592,
    "sensor_1": 23.5
}

client.publish("telemetry/data", json.dumps(data))
client.disconnect()
```

### Health Check

```bash
curl http://localhost:8080/health
```

**Respuesta:**
```json
{
  "status": "ok",
  "timestamp": "2025-12-05T10:30:00-03:00"
}
```

## 📁 Estructura del Proyecto

```
telemetria-endpoint-GOLANG/
├── cmd/
│   └── server/
│       └── main.go                 # Punto de entrada de la aplicación
├── internal/
│   ├── config/
│   │   └── config.go              # Gestión de configuración
│   ├── models/
│   │   └── telemetry.go           # Modelos de datos
│   ├── database/
│   │   ├── interface.go           # Interfaz de abstracción
│   │   ├── mysql/
│   │   │   ├── connection.go     # Conexión MySQL
│   │   │   └── repository.go     # Operaciones MySQL
│   │   └── redis/
│   │       ├── connection.go     # Conexión Redis
│   │       └── cache.go          # Operaciones de caché
│   ├── mqtt/
│   │   ├── client.go             # Cliente MQTT
│   │   └── handler.go            # Manejador de mensajes
│   ├── http/
│   │   ├── server.go             # Servidor HTTP
│   │   ├── handlers.go           # Manejadores de rutas
│   │   └── middleware.go         # Middlewares
│   ├── service/
│   │   ├── telemetry.go          # Lógica de negocio
│   │   ├── validation.go         # Validaciones
│   │   └── distance.go           # Cálculo de distancia
│   └── logger/
│       └── logger.go              # Sistema de logging
├── migrations/
│   └── mysql_schema.sql           # Esquema de base de datos
├── logs/                           # Directorio de logs (generado)
│   ├── app.log                    # Log de aplicación
│   ├── invalid_requests.log       # Peticiones inválidas
│   └── devices/                   # Logs por dispositivo
│       ├── DEVICE001.log
│       └── DEVICE002.log
├── .env.example                    # Plantilla de configuración
├── .env                            # Configuración (no versionado)
├── go.mod                          # Dependencias Go
├── go.sum                          # Checksums de dependencias
└── README.md                       # Esta documentación
```

## 🧩 Módulos

### `cmd/server/main.go`
Punto de entrada de la aplicación. Inicializa todos los componentes, gestiona el ciclo de vida y el graceful shutdown.

### `internal/config`
Gestión de configuración mediante variables de entorno. Soporta valores por defecto y validación.

### `internal/models`
Definición de estructuras de datos para telemetría, dispositivos, mediciones y errores.

### `internal/database`
**Abstracción de base de datos** que permite migrar fácilmente a otros motores SQL.

- **`interface.go`**: Define las interfaces `Repository` y `Cache`
- **`mysql/`**: Implementación para MySQL con pool de conexiones
- **`redis/`**: Implementación de caché con estructura hash y TTL

### `internal/http`
Servidor HTTP con framework Gin.

- **`server.go`**: Configuración del servidor y rutas
- **`handlers.go`**: Manejadores de endpoints
- **`middleware.go`**: Rate limiting y logging

### `internal/mqtt`
Cliente MQTT con auto-reconexión.

- **`client.go`**: Gestión de conexión MQTT
- **`handler.go`**: Procesamiento de mensajes

### `internal/service`
Lógica de negocio principal.

- **`telemetry.go`**: Procesamiento de datos, validación offline, gestión de caché
- **`validation.go`**: Validación de campos requeridos
- **`distance.go`**: Cálculo de distancia con fórmula de Haversine

### `internal/logger`
Sistema de logging estructurado.

- Logs de aplicación (info, warning, error)
- Logs por dispositivo (archivo separado por identificador)
- Logs de peticiones inválidas con IP de origen

## 🔄 Migración a Otras Bases de Datos

El proyecto utiliza una **interfaz de abstracción** que facilita la migración a otros motores SQL.

### PostgreSQL

1. **Instalar driver:**
```bash
go get github.com/lib/pq
```

2. **Crear implementación:**
```go
// internal/database/postgresql/connection.go
package postgresql

import (
    "database/sql"
    _ "github.com/lib/pq"
)

func NewConnection(cfg *config.PostgreSQLConfig) (*Connection, error) {
    dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database)
    
    db, err := sql.Open("postgres", dsn)
    // ... resto de la implementación
}
```

3. **Adaptar esquema:**
```sql
-- migrations/postgresql_schema.sql
CREATE TABLE equipos_telemetria (
    idTelemetria SERIAL PRIMARY KEY,
    Identificador VARCHAR(255) NOT NULL UNIQUE,
    Nombre VARCHAR(255) NOT NULL,
    UltimaConexion TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    TiempoFueraLinea TIME DEFAULT '00:00:00'
);
-- ... resto de las tablas
```

4. **Actualizar main.go:**
```go
import "github.com/tenshi98/telemetria-endpoint-GOLANG/internal/database/postgresql"

// Cambiar:
mysqlConn, err := mysql.NewConnection(&cfg.MySQL)
// Por:
pgConn, err := postgresql.NewConnection(&cfg.PostgreSQL)
```

### SQL Server

Similar al proceso de PostgreSQL, usando el driver `github.com/denisenkom/go-mssqldb`.

**Cambios en el esquema:**
- `AUTO_INCREMENT` → `IDENTITY(1,1)`
- `BIGINT UNSIGNED` → `BIGINT`
- Ajustar tipos de datos según SQL Server

### SQLite (para desarrollo/testing)

```bash
go get github.com/mattn/go-sqlite3
```

Ideal para pruebas locales sin servidor MySQL.

## ✅ Buenas Prácticas Implementadas

### Manejo de Errores
- Errores envueltos con contexto (`fmt.Errorf` con `%w`)
- Logging de todos los errores
- Respuestas HTTP apropiadas
- Registro en base de datos de errores de telemetría

### Timeouts y Delays
- Timeouts en conexiones de base de datos
- Timeouts en operaciones HTTP
- Delays configurables entre peticiones
- Context con timeout en operaciones MQTT

### Rate Limiting
- Algoritmo de token bucket
- Rate limiting por IP
- Configuración flexible (RPS y burst)
- Limpieza automática de clientes antiguos

### Modularización
- Separación clara de responsabilidades
- Paquetes internos bien definidos
- Interfaces para abstracción
- Código reutilizable

### Gestión de Conexiones
- Pool de conexiones MySQL configurable
- Pool de conexiones Redis
- Auto-reconexión MQTT
- Cierre ordenado de recursos (defer)

### Documentación
- Comentarios en código
- README completo
- Ejemplos de uso
- Guías de migración

### Logging
- Niveles de log (info, warning, error)
- Logs estructurados con timestamps
- Logs por dispositivo
- Logs de peticiones inválidas con IP
- Rotación manual de logs (crear script si es necesario)

### Seguridad
- Validación de entrada
- Rate limiting
- Configuración mediante variables de entorno
- No hardcodear credenciales

## 📝 Logs

### Ubicación de logs

- **Aplicación**: `./logs/app.log`
- **Peticiones inválidas**: `./logs/invalid_requests.log`
- **Por dispositivo**: `./logs/devices/{IDENTIFICADOR}.log`

### Ejemplo de log de aplicación

```
2025/12/05 10:30:15 [INFO] Starting Telemetry Endpoint Server...
2025/12/05 10:30:15 [INFO] MySQL connection established
2025/12/05 10:30:15 [INFO] Redis connection established
2025/12/05 10:30:15 [INFO] Starting HTTP server on port 8080
2025/12/05 10:30:20 [INFO] POST /telemetry - Status: 200 - Duration: 15ms - IP: 192.168.1.100
```

### Ejemplo de log por dispositivo

```
2025/12/05 10:30:20 Identificador: DEVICE001, Latitud: -34.603722, Longitud: -58.381592, Sensor_1: 23.500000
```

### Ejemplo de log de peticiones inválidas

```
2025/12/05 10:30:25 IP: 192.168.1.105, Timestamp: 2025-12-05T10:30:25-03:00, Identificador: MISSING, Latitud: MISSING, Longitud: -58.381592, Errors: [identificador: El identificador es requerido, latitud: La latitud es requerida]
```

## 🐛 Solución de Problemas

### Error: "go: command not found"

Instalar Go:
```bash
sudo apt install golang-go
```

### Error de conexión a MySQL

Verificar que MySQL esté corriendo:
```bash
sudo systemctl status mysql
```

Verificar credenciales en `.env`

### Error de conexión a Redis

Verificar que Redis esté corriendo:
```bash
sudo systemctl status redis-server
redis-cli ping
```

### MQTT no conecta

Verificar broker MQTT:
```bash
sudo systemctl status mosquitto
```

Probar conexión:
```bash
mosquitto_sub -t "test" -v
```

### Rate limit muy restrictivo

Ajustar en `.env`:
```env
RATE_LIMIT_RPS=1000.0
RATE_LIMIT_BURST=2000
```

