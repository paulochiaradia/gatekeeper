package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
	"github.com/paulochiaradia/gatekeeper/internal/config"
	"github.com/pressly/goose/v3"
)

// Connect inicializa o banco e aplica as migrations pendentes
func Connect() (*sql.DB, error) {
	cfg := config.GetConfig()

	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, err
	}

	// Testa a conexão real (Ping)
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Configura e executa as migrations
	goose.SetDialect("sqlite3")
	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("Falha ao rodar migrations: %v", err)
	}

	log.Println("Banco de dados conectado e migrations aplicadas.")
	return db, nil
}
