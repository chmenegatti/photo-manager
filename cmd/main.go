package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin" // Vamos usar o Gin como framework
)

func main() {
	// Inicializa o roteador do Gin
	router := gin.Default()

	// Define uma rota simples para testar
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Inicia o servidor HTTP
	port := "8080"
	fmt.Printf("Servidor iniciado na porta %s\n", port)
	log.Fatal(router.Run(":" + port)) // Inicia o servidor na porta especificada
}
