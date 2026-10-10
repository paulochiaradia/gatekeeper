package security_test

import (
	"testing"

	"github.com/paulochiaradia/gatekeeper/internal/core/security"
	"github.com/stretchr/testify/assert"
)

func TestECDHHandshake_Success(t *testing.T) {
	// 1. Simula a inicialização do Servidor Go
	serverManager, err := security.NewECDHManager()
	assert.NoError(t, err, "Servidor deve gerar chaves sem erro")
	serverPubKeyBase64 := serverManager.GetPublicKeyBase64()

	// 2. Simula a inicialização do ESP32 (Cliente)
	clientManager, err := security.NewECDHManager()
	assert.NoError(t, err, "Cliente deve gerar chaves sem erro")
	clientPubKeyBase64 := clientManager.GetPublicKeyBase64()

	// 3. A TROCA DE CHAVES (Handshake)

	// Servidor calcula o segredo usando a chave pública do cliente
	serverSharedSecret, err := serverManager.ComputeSharedSecret(clientPubKeyBase64)
	assert.NoError(t, err, "Servidor deve computar o segredo sem erro")

	// Cliente calcula o segredo usando a chave pública do servidor
	clientSharedSecret, err := clientManager.ComputeSharedSecret(serverPubKeyBase64)
	assert.NoError(t, err, "Cliente deve computar o segredo sem erro")

	// 4. A PROVA MATEMÁTICA: Os dois segredos devem ser EXATAMENTE iguais
	assert.Equal(t, serverSharedSecret, clientSharedSecret, "OS SEGREDOS COMPARTILHADOS NÃO BATEM! Criptografia falhou.")

	// Garante que o segredo gerado tem o tamanho correto para o AES-256 (32 bytes)
	assert.Len(t, serverSharedSecret, 32, "O segredo compartilhado deve ter 32 bytes para o AES-256")
}

func TestECDHHandshake_InvalidBase64(t *testing.T) {
	serverManager, _ := security.NewECDHManager()

	// Simula um ataque ou erro de rede mandando lixo no lugar da chave pública
	lixoBase64 := "IssoNaoEUmBase64Valido!@#"

	_, err := serverManager.ComputeSharedSecret(lixoBase64)

	// O servidor DEVE retornar um erro e não deve "capotar" (panic)
	assert.Error(t, err, "Servidor deveria rejeitar um Base64 inválido")
	assert.Contains(t, err.Error(), "formato base64 inválido")
}
