package ports

import "github.com/paulochiaradia/gatekeeper/internal/core/domain"

// UserRepository define as regras que qualquer banco de dados de usuários deve seguir.
type UserRepository interface {
	GetUserByUID(uid string) (*domain.User, error)
	CreateUser(user *domain.User) error
}

// AccessLogRepository define como os logs devem ser persistidos.
type AccessLogRepository interface {
	SaveLog(log *domain.AccessLog) error
}
