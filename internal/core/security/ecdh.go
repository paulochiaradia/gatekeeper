package security

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"errors"
)

// ECDHManager encapsula a lógica de geração de chaves e derivação de segredos
type ECDHManager struct {
	privateKey *ecdh.PrivateKey
	publicKey  *ecdh.PublicKey
}

// NewECDHManager gera as chaves do Servidor assim que a aplicação sobe
func NewECDHManager() (*ECDHManager, error) {
	// Usamos a Curva P-256, que é amplamente suportada em microcontroladores
	curve := ecdh.P256()

	// Gera a chave privada usando o gerador de entropia randômica do sistema operacional
	privKey, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &ECDHManager{
		privateKey: privKey,
		publicKey:  privKey.PublicKey(),
	}, nil
}

// GetPublicKeyBase64 exporta a chave pública do Servidor em Base64 para enviarmos ao ESP32
func (m *ECDHManager) GetPublicKeyBase64() string {
	pubBytes := m.publicKey.Bytes()
	return base64.StdEncoding.EncodeToString(pubBytes)
}

// ComputeSharedSecret recebe a chave pública do ESP32 (em Base64) e gera a senha mútua (AES Key)
func (m *ECDHManager) ComputeSharedSecret(clientPubKeyBase64 string) ([]byte, error) {
	// 1. Decodifica o texto em Base64 para Bytes puros
	clientPubBytes, err := base64.StdEncoding.DecodeString(clientPubKeyBase64)
	if err != nil {
		return nil, errors.New("formato base64 inválido")
	}

	// 2. Reconstrói a Chave Pública do ESP32 a partir dos bytes
	clientPubKey, err := ecdh.P256().NewPublicKey(clientPubBytes)
	if err != nil {
		return nil, errors.New("chave pública do cliente inválida")
	}

	// 3. Calcula o segredo compartilhado usando ECDH
	sharedSecret, err := m.privateKey.ECDH(clientPubKey)
	if err != nil {
		return nil, err
	}

	// Esse sharedSecret será a chave do AES-GCM (normalmente 32 bytes)
	return sharedSecret, nil
}
