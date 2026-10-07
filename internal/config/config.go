package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath       string
	MQTTBroker   string
	MQTTUsername string
	MQTTPassword string
}

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		_ = godotenv.Load() // Tenta carregar o .env

		instance = &Config{
			DBPath:       getEnvOrDefault("DB_PATH", "./gatekeeper.db"),
			MQTTBroker:   getEnvOrDefault("MQTT_BROKER", "tcp://localhost:1883"),
			MQTTUsername: getEnvOrDefault("MQTT_USERNAME", "gatekeeper"),
			MQTTPassword: getEnvOrDefault("MQTT_PASSWORD", ""),
		}
		log.Println("Configurações carregadas com segurança.")
	})
	return instance
}

func getEnvOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
