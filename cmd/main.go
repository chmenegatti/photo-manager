package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"photo-manager/internal/database" // Importa nosso pacote de database

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // Para carregar variáveis de ambiente
)

func main() {
	// Carrega as variáveis de ambiente do arquivo .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Atenção: Nenhum arquivo .env encontrado. Usando variáveis de ambiente do sistema.")
	}

	// Obtém a porta do ambiente ou usa 8080 como padrão
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// Obtém o caminho do banco de dados do ambiente
	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		dbPath = "./data/photo_manager.db" // Caminho padrão se não estiver no .env
		log.Printf("DATABASE_URL não configurado. Usando padrão: %s\n", dbPath)
	}

	// Garante que o diretório 'data' exista para o SQLite
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Falha ao criar diretório de dados '%s': %v", dataDir, err)
	}

	// Inicializa a conexão com o banco de dados
	database.InitDB(dbPath)

	// Inicializa o roteador do Gin
	router := gin.Default()

	// Define uma rota simples para testar
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Inicia o servidor HTTP
	fmt.Printf("Servidor iniciado na porta %s\n", port)
	log.Fatal(router.Run(":" + port)) // Inicia o servidor na porta especificada
}
