package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB // Variável global para a instância do banco de dados

// InitDB inicializa a conexão com o banco de dados e realiza as migrações.
func InitDB(databasePath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Falha ao conectar ao banco de dados: %v", err)
	}

	// Migração automática do schema
	err = DB.AutoMigrate(&Photo{}, &Album{}, &AlbumPhoto{})
	if err != nil {
		log.Fatalf("Falha ao migrar o schema do banco de dados: %v", err)
	}

	log.Println("Conexão com o banco de dados estabelecida e migrações executadas com sucesso!")
}
