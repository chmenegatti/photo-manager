package api

import (
	"fmt"
	"log"
	"net/http"
	"photo-manager/internal/service"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PhotoHandler gerencia as requisições HTTP para fotos.
type PhotoHandler struct {
	PhotoService *service.PhotoService
}

// NewPhotoHandler cria uma nova instância de PhotoHandler.
func NewPhotoHandler(s *service.PhotoService) *PhotoHandler {
	return &PhotoHandler{
		PhotoService: s,
	}
}

// UploadPhotoHandler lida com o upload de uma ou múltiplas fotos.
func (h *PhotoHandler) UploadPhotoHandler(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Não foi possível ler o formulário multipart: %v", err)})
		return
	}

	files := form.File["photos"] // Nome do campo do input type="file" no HTML

	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum arquivo 'photos' encontrado no formulário."})
		return
	}

	uploadedPhotos := []map[string]string{}
	errors := []map[string]string{}

	for _, file := range files {
		// Adicionar validação de MIME type e tamanho máximo aqui!
		// Exemplo básico de validação de MIME type:
		if file.Header.Get("Content-Type") != "image/jpeg" && file.Header.Get("Content-Type") != "image/png" {
			errors = append(errors, map[string]string{"filename": file.Filename, "error": "Tipo de arquivo não permitido. Apenas JPG/PNG."})
			continue
		}

		// Limite de 10MB por arquivo
		const maxUploadSize = 10 << 20 // 10 MB
		if file.Size > maxUploadSize {
			errors = append(errors, map[string]string{"filename": file.Filename, "error": fmt.Sprintf("Tamanho do arquivo excede o limite de %dMB", maxUploadSize/(1<<20))})
			continue
		}

		photo, err := h.PhotoService.UploadPhoto(file)
		if err != nil {
			log.Printf("Erro ao processar o upload da foto '%s': %v\n", file.Filename, err)
			errors = append(errors, map[string]string{"filename": file.Filename, "error": err.Error()})
		} else {
			uploadedPhotos = append(uploadedPhotos, map[string]string{
				"id":        fmt.Sprintf("%d", photo.ID),
				"filename":  photo.Filename,
				"stored_at": photo.StoredPath,
				"exif_date": func() string {
					if photo.ExifDate != nil {
						return photo.ExifDate.Format(time.RFC3339)
					}
					return ""
				}(),
			})
		}
	}

	if len(errors) > 0 {
		c.JSON(http.StatusMultiStatus, gin.H{
			"message":  "Algumas fotos foram processadas com erros.",
			"uploaded": uploadedPhotos,
			"errors":   errors,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message":  "Uploads processados com sucesso!",
			"uploaded": uploadedPhotos,
		})
	}
}

// GetPhotosHandler lida com a busca e listagem de fotos com filtros.
func (h *PhotoHandler) GetPhotosHandler(c *gin.Context) {
	var filter service.PhotoFilter

	// Extrai parâmetros da query string
	if yearStr := c.Query("year"); yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ano inválido."})
			return
		}
		filter.Year = year
	}
	if monthStr := c.Query("month"); monthStr != "" {
		month, err := strconv.Atoi(monthStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Mês inválido."})
			return
		}
		filter.Month = month
	}
	filter.Filename = c.Query("filename")
	filter.Tag = c.Query("tag")

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Limite inválido."})
			return
		}
		filter.Limit = limit
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Offset inválido."})
			return
		}
		filter.Offset = offset
	}
	filter.OrderBy = c.Query("order_by")

	photos, err := h.PhotoService.GetPhotos(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao buscar fotos: %v", err)})
		return
	}

	// Transformar as fotos para um formato de resposta mais amigável, se necessário
	// Por exemplo, formatar datas, incluir URLs de acesso, etc.
	responsePhotos := []gin.H{}
	for _, photo := range photos {
		responsePhotos = append(responsePhotos, gin.H{
			"id":          photo.ID,
			"filename":    photo.Filename,
			"stored_path": photo.StoredPath,
			"upload_date": photo.UploadDate.Format(time.RFC3339),
			"exif_date": func() string {
				if photo.ExifDate != nil {
					return photo.ExifDate.Format(time.RFC3339)
				}
				return ""
			}(),
			"hash":           photo.Hash,
			"file_size":      photo.FileSize,
			"mime_type":      photo.MimeType,
			"width":          photo.Width,
			"height":         photo.Height,
			"description":    photo.Description,
			"tags":           photo.Tags,
			"thumbnail_path": photo.ThumbnailPath, // Incluir se houver miniaturas
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": responsePhotos})
}

// GetPhotosTimelineHandler retorna fotos organizadas por ano e mês.
func (h *PhotoHandler) GetPhotosTimelineHandler(c *gin.Context) {
	limitPerMonthStr := c.DefaultQuery("limit_per_month", "0") // Default 0 means no limit
	limitPerMonth, err := strconv.Atoi(limitPerMonthStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Limite por mês inválido."})
		return
	}

	timeline, err := h.PhotoService.GetPhotosByTimeline(limitPerMonth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Erro ao buscar linha do tempo: %v", err)})
		return
	}

	// Transformar para um formato de resposta mais amigável, especialmente para datas e aninhamento
	responseTimeline := gin.H{}
	for year, months := range timeline {
		yearStr := strconv.Itoa(year)
		responseTimeline[yearStr] = gin.H{}
		for month, photos := range months {
			monthStr := fmt.Sprintf("%02d", month) // Formatar mês com dois dígitos
			photoList := []gin.H{}
			for _, photo := range photos {
				photoList = append(photoList, gin.H{
					"id":          photo.ID,
					"filename":    photo.Filename,
					"stored_path": photo.StoredPath,
					"upload_date": photo.UploadDate.Format(time.RFC3339),
					"exif_date": func() string {
						if photo.ExifDate != nil {
							return photo.ExifDate.Format(time.RFC3339)
						}
						return ""
					}(),
					"hash":           photo.Hash,
					"file_size":      photo.FileSize,
					"mime_type":      photo.MimeType,
					"width":          photo.Width,
					"height":         photo.Height,
					"description":    photo.Description,
					"tags":           photo.Tags,
					"thumbnail_path": photo.ThumbnailPath,
				})
			}
			responseTimeline[yearStr].(gin.H)[monthStr] = photoList
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": responseTimeline})
}
