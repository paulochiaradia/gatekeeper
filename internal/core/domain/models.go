package domain

import "time"

// User representa a identidade de quem tenta acessar a porta.
type User struct {
	ID        int
	UID       string // Identificador do cartão RFID (MFRC522)
	PublicKey string // Chave pública ECDH do usuário para a criptografia futura
	Name      string
	IsActive  bool
}

// AccessLog representa o registro inalterável de uma tentativa de acesso.
type AccessLog struct {
	ID        int
	UserUID   string
	Status    string // Ex: "GRANTED" (Concedido) ou "DENIED" (Negado)
	Timestamp time.Time
}
