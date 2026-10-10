package ports

import (
	"context"

	"github.com/paulochiaradia/gatekeeper/internal/core/domain"
)

// UserRepository define as regras que qualquer banco de dados de usuários deve seguir.
type UserRepository interface {
	GetUserByUID(ctx context.Context, uid string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error
}

// AccessLogRepository define como os logs devem ser persistidos.
type AccessLogRepository interface {
	SaveLog(ctx context.Context, log *domain.AccessLog) error
}
