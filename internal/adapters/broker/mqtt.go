package broker

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/paulochiaradia/gatekeeper/internal/core/ports"
)

// MQTTAdapter traduz as mensagens da rede para os casos de uso de negócio
type MQTTAdapter struct {
	client  mqtt.Client
	usecase ports.AccessUseCase
}

// NewMQTTAdapter configura e inicializa a conexão com o broker Mosquitto
func NewMQTTAdapter(brokerURI, clientID, username, password string, uc ports.AccessUseCase) (*MQTTAdapter, error) {
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
	}, nil
}

// StartListening assina o tópico e define a função de callback
func (m *MQTTAdapter) StartListening(requestTopic string) error {
	token := m.client.Subscribe(requestTopic, 1, m.messageHandler)
	token.Wait()

	if token.Error() != nil {
		return token.Error()
	}

	log.Printf("[MQTT] Inscrito com sucesso no tópico: %s", requestTopic)
	return nil
}

// messageHandler é acionado de forma assíncrona sempre que o ESP32 manda mensagem
func (m *MQTTAdapter) messageHandler(client mqtt.Client, msg mqtt.Message) {
	payload := string(msg.Payload())
	log.Printf("[MQTT EVENT] Mensagem recebida -> Tópico: %s | Payload: %s", msg.Topic(), payload)

	// Repassa para a regra de negócio
	granted, err := m.usecase.ProcessAccessRequest(payload)

	// Prepara a resposta (Publish) de volta para o ESP32
	responseTopic := "security/door/response"
	var responseMsg string

	if err != nil || !granted {
		log.Printf("[MQTT ADAPTER] Enviando comando DENIED")
		responseMsg = "DENIED"
	} else {
		log.Printf("[MQTT ADAPTER] Enviando comando GRANTED")
		responseMsg = "GRANTED"
	}

	// Publica a resposta de volta (QoS 1)
	token := client.Publish(responseTopic, 1, false, responseMsg)
	token.Wait()
}

// Disconnect desliga o adaptador graciosamente
func (m *MQTTAdapter) Disconnect() {
	m.client.Disconnect(250)
	log.Println("[MQTT] Desconectado do broker.")
}
