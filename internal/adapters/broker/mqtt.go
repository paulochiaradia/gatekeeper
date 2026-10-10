package broker

import (
	"context"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/paulochiaradia/gatekeeper/internal/core/ports"
	"github.com/paulochiaradia/gatekeeper/internal/core/security"
)

// MQTTAdapter é o adaptador que conecta o nosso Servidor Go ao Broker MQTT (Mosquitto)
type MQTTAdapter struct {
	client  mqtt.Client
	usecase ports.AccessUseCase
	crypto  *security.ECDHManager
}

// NewMQTTAdapter cria uma nova instância do adaptador MQTT
func NewMQTTAdapter(brokerURI, clientID, username, password string, uc ports.AccessUseCase, crypto *security.ECDHManager) (*MQTTAdapter, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURI)
	opts.SetClientID(clientID)

	if username != "" {
		opts.SetUsername(username)
		opts.SetPassword(password)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if token.Error() != nil {
		return nil, token.Error()
	}

	return &MQTTAdapter{
		client:  client,
		usecase: uc,
		crypto:  crypto, // Guardamos a referência
	}, nil
}

// StartListening agora assina DOIS tópicos
func (m *MQTTAdapter) StartListening(requestTopic, handshakeTopic string) error {
	// Assina o tópico de requisição de acesso (Porta)
	token1 := m.client.Subscribe(requestTopic, 1, m.doorRequestHandler)
	token1.Wait()
	if token1.Error() != nil {
		return token1.Error()
	}

	// Assina o tópico de troca de chaves (Handshake)
	token2 := m.client.Subscribe(handshakeTopic, 1, m.handshakeHandler)
	token2.Wait()
	if token2.Error() != nil {
		return token2.Error()
	}

	log.Printf("[MQTT] Inscrito nos tópicos:\n -> %s\n -> %s", requestTopic, handshakeTopic)
	return nil
}

// handshakeHandler processa a chave pública do ESP32 e devolve a do Servidor
func (m *MQTTAdapter) handshakeHandler(client mqtt.Client, msg mqtt.Message) {
	esp32PubKeyBase64 := string(msg.Payload())
	log.Printf("[SECURITY] Pedido de Handshake recebido. Chave do ESP32: %s...", esp32PubKeyBase64[:10])

	// Tenta calcular o Segredo Compartilhado
	sharedSecret, err := m.crypto.ComputeSharedSecret(esp32PubKeyBase64)
	if err != nil {
		log.Printf("[SECURITY ERROR] Falha ao processar chave do ESP32: %v", err)
		return
	}

	// SUCESSO! Na Fase 3 nós usaremos esse segredo para inicializar o AES-GCM
	log.Printf("[SECURITY SUCCESS] Segredo Compartilhado (Shared Secret) gerado com sucesso! Tamanho: %d bytes.", len(sharedSecret))

	// Responde com a Chave Pública do Servidor Go
	responseTopic := "security/handshake/response"
	serverPubKey := m.crypto.GetPublicKeyBase64()

	token := client.Publish(responseTopic, 1, false, serverPubKey)
	token.Wait()
	log.Printf("[SECURITY] Chave Pública do Servidor enviada de volta ao ESP32.")
}

// doorRequestHandler é o nosso handler antigo (apenas renomeado para ficar claro)
func (m *MQTTAdapter) doorRequestHandler(client mqtt.Client, msg mqtt.Message) {
	payload := string(msg.Payload())
	log.Printf("[MQTT EVENT] Pedido de porta -> Payload: %s", payload)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // Libera os recursos da memória ao sair da função

	// Passamos o escudo protetor (ctx) para a regra de negócio
	granted, err := m.usecase.ProcessAccessRequest(ctx, payload)

	responseTopic := "security/door/response"
	var responseMsg string

	if err != nil || !granted {
		responseMsg = "DENIED"
	} else {
		responseMsg = "GRANTED"
	}

	client.Publish(responseTopic, 1, false, responseMsg).Wait()
}

func (m *MQTTAdapter) Disconnect() {
	m.client.Disconnect(250)
	log.Println("[MQTT] Desconectado do broker.")
}
