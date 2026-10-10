package services

import (
	"context"
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
func (s *AccessService) ProcessAccessRequest(ctx context.Context, uid string) (bool, error) {
	// 1. Busca o usuário pelo UID do cartão
	user, err := s.userRepo.GetUserByUID(ctx, uid)

	if err != nil {
		if errors.Is(err, errors.New("usuário não encontrado")) {
			// Registra a tentativa de acesso com status "falha"
			accessLog := &domain.AccessLog{
				UserUID:   uid,
				Timestamp: time.Now(),
				Status:    "DENIED_NOT_FOUND",
			}
			_ = s.logRepo.SaveLog(ctx, accessLog)
			log.Printf("[AUDIT] Acesso negado: UID %s não encontrado.", uid)
			return false, errors.New("acesso negado")
		}
		return false, err
	}

	// 2. Regra: Usuário não existe
	if user == nil {
		return false, errors.New("usuário não encontrado")
	}

	// 3. Regra: Usuário existe, mas está inativo/bloqueado
	if !user.IsActive {
		accessLog := &domain.AccessLog{
			UserUID:   uid,
			Timestamp: time.Now(),
			Status:    "DENIED_INACTIVE",
		}
		_ = s.logRepo.SaveLog(ctx, accessLog)
		log.Printf("[AUDIT] Acesso negado: Usuário %s (%s) inativo.", user.Name, uid)
		return false, errors.New("usuário inativo")
	}

	// 4. Regra: Sucesso
	accessLog := &domain.AccessLog{
		UserUID:   uid,
		Timestamp: time.Now(),
		Status:    "GRANTED",
	}
	_ = s.logRepo.SaveLog(ctx, accessLog)
	log.Printf("[AUDIT] Acesso concedido: Bem-vindo, %s.", user.Name)

	return true, nil
}
