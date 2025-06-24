package api

import (
	"fmt"
	"log"
	"net/http"
	"photo-manager/internal/service"
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
