package services

import (
	"errors"

	"log"
	"time"

	"github.com/paulochiaradia/gatekeeper/internal/core/domain"
	"github.com/paulochiaradia/gatekeeper/internal/core/ports"
)

// AccessService é o orquestrador das regras de negócio
type AccessService struct {
	userRepo ports.UserRepository
	logRepo  ports.AccessLogRepository
}

// NewAccessService injeta as dependências necessárias para o serviço funcionar
func NewAccessService(ur ports.UserRepository, lr ports.AccessLogRepository) *AccessService {
	return &AccessService{
		userRepo: ur,
		logRepo:  lr,
	}
}

// ProcessAccessRequest valida o crachá e registra a tentativa no banco
func (s *AccessService) ProcessAccessRequest(uid string) (bool, error) {
	// 1. Busca o usuário pelo UID do cartão
	user, err := s.userRepo.GetUserByUID(uid)

	accessLog := &domain.AccessLog{
		UserUID:   uid,
		Timestamp: time.Now(),
	}

	// 2. Regra: Usuário não existe
	if err != nil {
		accessLog.Status = "DENIED_NOT_FOUND"
		_ = s.logRepo.SaveLog(accessLog) // O underline ignora o erro do log para não travar a execução
		log.Printf("[AUDIT] Acesso negado: UID %s não encontrado.", uid)
		return false, errors.New("acesso negado")
	}

	// 3. Regra: Usuário existe, mas está inativo/bloqueado
	if !user.IsActive {
		accessLog.Status = "DENIED_INACTIVE"
		_ = s.logRepo.SaveLog(accessLog)
		log.Printf("[AUDIT] Acesso negado: Usuário %s (%s) inativo.", user.Name, uid)
		return false, errors.New("usuário inativo")
	}

	// 4. Regra: Sucesso
	accessLog.Status = "GRANTED"
	_ = s.logRepo.SaveLog(accessLog)
	log.Printf("[AUDIT] Acesso concedido: Bem-vindo, %s.", user.Name)

	return true, nil
}
