package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"photo-manager/internal/api"
	"photo-manager/internal/database"
	"photo-manager/internal/service"
	"photo-manager/internal/storage" // Importa nosso pacote de storage

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

	// Obtém o caminho para armazenamento das fotos
	photoStoragePath := os.Getenv("PHOTO_STORAGE_PATH")
	if photoStoragePath == "" {
		photoStoragePath = "./data/photos" // Caminho padrão
		log.Printf("PHOTO_STORAGE_PATH não configurado. Usando padrão: %s\n", photoStoragePath)
	}
	// Garante que o diretório de armazenamento de fotos exista
	if err := os.MkdirAll(photoStoragePath, 0755); err != nil {
		log.Fatalf("Falha ao criar diretório de armazenamento de fotos '%s': %v", photoStoragePath, err)
	}

	// Inicializa a conexão com o banco de dados
	database.InitDB(dbPath)

	// Inicializa o gerenciador de arquivos
	fileManager := storage.NewFileManager(photoStoragePath)

	// Inicializa o serviço de fotos
	photoService := service.NewPhotoService(database.DB, fileManager)

	// Inicializa o handler da API de fotos
	photoHandler := api.NewPhotoHandler(photoService)

	// Inicializa o roteador do Gin
	router := gin.Default()

	// Define uma rota simples para testar
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Rota para upload de fotos
	// Cuidado: Gin tem um limite de tamanho de corpo padrão. Para uploads maiores, configure MaxMultipartMemory
	// router.MaxMultipartMemory = 8 << 20 // 8MB - padrão. Aumente se necessário, ex: 64 << 20 (64MB)
	router.POST("/upload", photoHandler.UploadPhotoHandler)

	// Novas rotas para busca e linha do tempo
	router.GET("/photos", photoHandler.GetPhotosHandler)
	router.GET("/photos/timeline", photoHandler.GetPhotosTimelineHandler)

	// Inicia o servidor HTTP
	fmt.Printf("Servidor iniciado na porta %s\n", port)
	log.Fatal(router.Run(":" + port)) // Inicia o servidor na porta especificada
}
