package repository_test

import (
	"database/sql"
	"testing"

	"github.com/paulochiaradia/gatekeeper/internal/adapters/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func setupInMemoryDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err, "Deve conectar no banco em memória")

	query := `
	CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uid TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		is_active INTEGER DEFAULT 1,
		public_key TEXT DEFAULT ''
	);`

	_, err = db.Exec(query)
	require.NoError(t, err, "Deve criar a tabela de usuários sem erro")

	return db
}

// TestSQLiteRepository_GetUserByUID_Success verifica se o método GetUserByUID retorna corretamente um usuário existente no banco.
func TestSQLiteRepository_GetUserByUID_Success(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	_, err := db.Exec("INSERT INTO users (uid, name, is_active, public_key) VALUES ('TAG-MESTRE', 'Paulo Chiaradia', 1, '')")
	require.NoError(t, err)

	repo := repository.NewSQLiteRepository(db)

	user, err := repo.GetUserByUID(t.Context(), "TAG-MESTRE")

	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "TAG-MESTRE", user.UID)
	assert.Equal(t, "Paulo Chiaradia", user.Name)
	assert.True(t, user.IsActive)
	assert.Equal(t, "", user.PublicKey)
}

// TestSQLiteRepository_GetUserByUID_NotFound verifica se o método GetUserByUID retorna um erro quando o usuário não é encontrado.
func TestSQLiteRepository_GetUserByUID_NotFound(t *testing.T) {
	db := setupInMemoryDB(t)
	defer db.Close()

	repo := repository.NewSQLiteRepository(db)

	user, err := repo.GetUserByUID(t.Context(), "TAG-FANTASMA")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "não encontrado")
}
