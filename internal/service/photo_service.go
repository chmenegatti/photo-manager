package service

import (
	"crypto/md5" // Ou sha256, para um hash mais robusto
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"photo-manager/internal/database"
	"photo-manager/internal/exif"
	"photo-manager/internal/storage"
	"time"

	"gorm.io/gorm"
)

// PhotoService define a interface para operações de foto.
type PhotoService struct {
	DB          *gorm.DB
	FileManager *storage.FileManager
}

// NewPhotoService cria uma nova instância de PhotoService.
func NewPhotoService(db *gorm.DB, fm *storage.FileManager) *PhotoService {
	return &PhotoService{
		DB:          db,
		FileManager: fm,
	}
}

// UploadPhoto processa o upload de uma foto, extrai metadados e a salva.
func (s *PhotoService) UploadPhoto(file *multipart.FileHeader) (*database.Photo, error) {
	uploadDate := time.Now()

	// 1. Salva o arquivo temporariamente para extração EXIF e hash
	// Criar um diretório temporário ou usar um sistema de fluxo de dados mais eficiente para arquivos grandes.
	// Por simplicidade, vamos salvar em um local temporário no disco.
	tempDir := filepath.Join(os.TempDir(), "photo-manager-temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("não foi possível criar diretório temporário: %w", err)
	}

	tempFilePath := filepath.Join(tempDir, file.Filename)
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("não foi possível abrir o arquivo enviado para processamento: %w", err)
	}
	defer src.Close()

	dstTemp, err := os.Create(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("não foi possível criar arquivo temporário: %w", err)
	}
	defer dstTemp.Close()
	defer os.Remove(tempFilePath) // Garante que o arquivo temporário seja removido

	_, err = io.Copy(dstTemp, src)
	if err != nil {
		return nil, fmt.Errorf("não foi possível copiar o arquivo para o temporário: %w", err)
	}
	dstTemp.Close() // Fecha o arquivo para garantir que todos os dados foram gravados antes de ler

	// 2. Extrai metadados EXIF
	exifData, err := exif.ExtractExifData(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao extrair dados EXIF: %w", err)
	}

	// Determina a data de organização (EXIF como preferência, senão data de upload)
	var photoOrganizeDate time.Time
	var exifDateTime *time.Time
	if exifData != nil && exifData.DateTime != nil {
		photoOrganizeDate = *exifData.DateTime
		exifDateTime = exifData.DateTime
	} else {
		photoOrganizeDate = uploadDate
	}

	// 3. Calcula o hash da foto (MD5 por simplicidade, SHA256 é mais robusto)
	hash, err := calculateMD5Hash(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("não foi possível calcular o hash da foto: %w", err)
	}

	// 4. Verifica duplicatas (futura funcionalidade - por enquanto, apenas um log)
	var existingPhoto database.Photo
	result := s.DB.Where("hash = ?", hash).First(&existingPhoto)
	if result.Error == nil {
		// Foto duplicada encontrada
		return &existingPhoto, fmt.Errorf("foto duplicada detectada (hash: %s, caminho existente: %s)", hash, existingPhoto.StoredPath)
	} else if result.Error != gorm.ErrRecordNotFound {
		// Erro real do banco de dados
		return nil, fmt.Errorf("erro ao verificar duplicatas: %w", result.Error)
	}

	// 5. Salva a foto no sistema de arquivos na estrutura ano/mês
	storedPath, err := s.FileManager.SavePhoto(file, photoOrganizeDate)
	if err != nil {
		return nil, fmt.Errorf("não foi possível salvar a foto no armazenamento: %w", err)
	}

	// 6. Preenche os metadados da foto
	photo := database.Photo{
		Filename:   file.Filename,
		StoredPath: storedPath,
		UploadDate: uploadDate,
		ExifDate:   exifDateTime,
		Hash:       hash,
		FileSize:   file.Size,
		MimeType:   file.Header.Get("Content-Type"), // Tipo MIME do upload
		// Largura e Altura podem ser extraídas com outra biblioteca de imagem se necessário (image/jpeg, etc.)
		// Por enquanto, deixamos em 0
		Width:  0,
		Height: 0,
	}

	// 7. Salva os metadados da foto no banco de dados
	if result := s.DB.Create(&photo); result.Error != nil {
		// Se falhar a gravação no DB, tentar remover o arquivo salvo para evitar lixo
		os.Remove(storedPath)
		return nil, fmt.Errorf("não foi possível salvar os metadados da foto no banco de dados: %w", result.Error)
	}

	return &photo, nil
}

// calculateMD5Hash calcula o hash MD5 de um arquivo.
func calculateMD5Hash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("não foi possível abrir o arquivo para hash: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("não foi possível calcular o hash do arquivo: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

type PhotoFilter struct {
	Year     int
	Month    int
	Filename string
	Tag      string
	Offset   int
	Limit    int
	OrderBy  string // Campo para ordenação (ex: "exif_date DESC", "upload_date ASC")
}

// GetPhotos busca fotos com base nos filtros fornecidos.
func (s *PhotoService) GetPhotos(filter PhotoFilter) ([]database.Photo, error) {
	query := s.DB.Model(&database.Photo{})

	if filter.Year != 0 {
		// Filtra por ano (tanto EXIF quanto UploadDate)
		// SQLite não tem funções DATE_PART, então usamos BETWEEN para o início e fim do ano.
		startDate := time.Date(filter.Year, time.January, 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(1, 0, 0).Add(-time.Nanosecond) // Fim do ano
		query = query.Where("(exif_date BETWEEN ? AND ?) OR (upload_date BETWEEN ? AND ?)", startDate, endDate, startDate, endDate)
	}

	if filter.Month != 0 {
		if filter.Year == 0 {
			// Se o mês for especificado sem o ano, é mais complexo e pode ser ineficiente em grandes bases.
			// Para SQLite, não há função nativa MONTH().
			// Uma solução robusta exigiria um campo extra para mês/ano ou uma consulta mais complexa.
			// Por simplicidade, vamos exigir o ano se o mês for filtrado por enquanto, ou ignorar se o ano não estiver presente.
			// Consideração futura: Adicionar colunas `photo_year` e `photo_month` na tabela `photos` para indexação mais eficiente.
			return nil, fmt.Errorf("filtrar por mês requer que o ano também seja especificado")
		}
		// Filtra por mês (assumindo que o ano já foi filtrado ou será)
		startDate := time.Date(filter.Year, time.Month(filter.Month), 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond) // Fim do mês
		query = query.Where("(exif_date BETWEEN ? AND ?) OR (upload_date BETWEEN ? AND ?)", startDate, endDate, startDate, endDate)
	}

	if filter.Filename != "" {
		query = query.Where("filename LIKE ?", "%"+filter.Filename+"%")
	}

	if filter.Tag != "" {
		// Busca por tags (assumindo tags separadas por vírgula)
		query = query.Where("tags LIKE ?", "%"+filter.Tag+"%")
	}

	// Ordenação
	if filter.OrderBy != "" {
		query = query.Order(filter.OrderBy)
	} else {
		// Ordem padrão: mais recente primeiro, priorizando EXIF, depois UploadDate
		query = query.Order("exif_date DESC").Order("upload_date DESC")
	}

	// Paginação
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var photos []database.Photo
	if result := query.Find(&photos); result.Error != nil {
		return nil, fmt.Errorf("erro ao buscar fotos: %w", result.Error)
	}

	return photos, nil
}

// GetPhotosByTimeline retorna fotos agrupadas por ano e mês para exibição em linha do tempo.
// Esta função pode ser otimizada para buscar apenas os anos/meses existentes primeiro.
func (s *PhotoService) GetPhotosByTimeline(limitPerMonth int) (map[int]map[int][]database.Photo, error) {
	// Poderíamos buscar todos os anos/meses distintos e depois buscar as fotos para cada um,
	// mas para simplicidade inicial, vamos buscar as fotos e agrupá-las em memória.
	// Para grandes volumes, seria melhor uma abordagem de paginação/streaming ou buscar apenas as fotos do "mês ativo".

	var photos []database.Photo
	// Pega todas as fotos, ordenadas para facilitar o agrupamento
	// A ordem preferencial é pela data EXIF, e depois pela data de upload
	result := s.DB.Order("exif_date DESC").Order("upload_date DESC").Find(&photos)
	if result.Error != nil {
		return nil, fmt.Errorf("erro ao buscar fotos para linha do tempo: %w", result.Error)
	}

	timeline := make(map[int]map[int][]database.Photo) // year -> month -> []Photo

	for _, photo := range photos {
		var dateToUse time.Time
		if photo.ExifDate != nil {
			dateToUse = *photo.ExifDate
		} else {
			dateToUse = photo.UploadDate
		}

		year := dateToUse.Year()
		month := int(dateToUse.Month())

		if _, ok := timeline[year]; !ok {
			timeline[year] = make(map[int][]database.Photo)
		}
		if _, ok := timeline[year][month]; !ok {
			timeline[year][month] = []database.Photo{}
		}

		// Adiciona a foto se o limite por mês ainda não foi atingido
		if limitPerMonth == 0 || len(timeline[year][month]) < limitPerMonth {
			timeline[year][month] = append(timeline[year][month], photo)
		}
	}

	return timeline, nil
}
