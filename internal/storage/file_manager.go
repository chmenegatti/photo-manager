package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// FileManager gerencia o armazenamento de arquivos.
type FileManager struct {
	BaseStoragePath string
}

// NewFileManager cria uma nova instância de FileManager.
func NewFileManager(basePath string) *FileManager {
	return &FileManager{
		BaseStoragePath: basePath,
	}
}

// SavePhoto salva um arquivo de foto no sistema de arquivos, organizando-o por ano e mês.
// Retorna o caminho completo onde a foto foi salva.
func (fm *FileManager) SavePhoto(file *multipart.FileHeader, photoDate time.Time) (string, error) {
	// Formato o caminho baseado na data da foto
	year := photoDate.Format("2006") // Ano completo (YYYY)
	month := photoDate.Format("01")  // Mês com dois dígitos (MM)

	// Cria o caminho completo para o diretório de destino
	targetDir := filepath.Join(fm.BaseStoragePath, year, month)

	// Garante que o diretório exista
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("não foi possível criar o diretório de destino '%s': %w", targetDir, err)
	}

	// Cria o caminho completo para o arquivo de destino
	filePath := filepath.Join(targetDir, file.Filename)

	// Abre o arquivo enviado
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("não foi possível abrir o arquivo enviado: %w", err)
	}
	defer src.Close()

	// Cria o arquivo de destino
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("não foi possível criar o arquivo de destino '%s': %w", filePath, err)
	}
	defer dst.Close()

	// Copia o conteúdo do arquivo enviado para o arquivo de destino
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("não foi possível copiar o arquivo para '%s': %w", filePath, err)
	}

	return filePath, nil
}
