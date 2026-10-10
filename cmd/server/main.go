package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/paulochiaradia/gatekeeper/internal/adapters/broker"
	"github.com/paulochiaradia/gatekeeper/internal/adapters/repository"
	"github.com/paulochiaradia/gatekeeper/internal/config"
	"github.com/paulochiaradia/gatekeeper/internal/core/security"
	"github.com/paulochiaradia/gatekeeper/internal/core/services"
	"github.com/paulochiaradia/gatekeeper/internal/infrastructure/database"
)

func main() {
	fmt.Println("=== Iniciando Gatekeeper Zero Trust  ===")

	cfg := config.GetConfig()

	cryptoManager, err := security.NewECDHManager()
	if err != nil {
		log.Fatalf("Falha crítica ao gerar chaves ECDH: %v", err)
	}
	log.Printf("[SECURITY] Chave Pública do Servidor (Base64): %s", cryptoManager.GetPublicKeyBase64())

	dbConnection, err := database.Connect()
	if err != nil {
		log.Fatalf("Erro crítico na inicialização do banco: %v", err)
	}
	defer dbConnection.Close()

	sqliteRepo := repository.NewSQLiteRepository(dbConnection)

	// Instanciando o Serviço de Negócio
	accessService := services.NewAccessService(sqliteRepo, sqliteRepo)

	// Instanciando o Adaptador MQTT (Substitua a senha pela que você criou no Mosquitto do Pi)
	mqttHandler, err := broker.NewMQTTAdapter(
		cfg.MQTTBroker,
		"gatekeeper-server-go",
		cfg.MQTTUsername,
		cfg.MQTTPassword,
		accessService,
		cryptoManager,
	)
	if err != nil {
		log.Fatalf("Falha ao conectar no MQTT: %v", err)
	}
	defer mqttHandler.Disconnect()

	// 6. Inicia a escuta dos dois tópicos
	err = mqttHandler.StartListening("security/door/request", "security/handshake/request")
	if err != nil {
		log.Fatalf("Falha ao assinar tópicos: %v", err)
	}

	log.Println("Servidor online e integrado. Pressione Ctrl+C para encerrar.")

	// Padrão de Desligamento Gracioso (Graceful Shutdown)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\nDesligando servidor graciosamente...")
}
