package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/paulochiaradia/gatekeeper/internal/core/domain"
)

type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository agora apenas recebe a conexão injetada
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// GetUserByUID busca um usuário no banco a partir do cartão
func (r *SQLiteRepository) GetUserByUID(ctx context.Context, uid string) (*domain.User, error) {
	var user domain.User
	query := `SELECT id, uid, public_key, name, is_active FROM users WHERE uid = ?`

	err := r.db.QueryRowContext(ctx, query, uid).Scan(&user.ID, &user.UID, &user.PublicKey, &user.Name, &user.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, err
	}
	return &user, nil
}

// CreateUser insere um novo crachá autorizado
func (r *SQLiteRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (uid, public_key, name, is_active) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, user.UID, user.PublicKey, user.Name, user.IsActive)
	return err
}

// SaveLog grava a tentativa de acesso para a auditoria
func (r *SQLiteRepository) SaveLog(ctx context.Context, log *domain.AccessLog) error {
	query := `INSERT INTO access_logs (user_uid, status) VALUES (?, ?)`
	_, err := r.db.ExecContext(ctx, query, log.UserUID, log.Status)
	return err
}
