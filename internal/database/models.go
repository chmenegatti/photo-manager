package database

import (
	"time"

	"gorm.io/gorm"
)

// Photo representa a estrutura de uma foto no banco de dados.
type Photo struct {
	gorm.Model                 // Inclui campos padrão como ID, CreatedAt, UpdatedAt, DeletedAt
	Filename      string       `gorm:"uniqueIndex;not null"` // Nome original do arquivo
	StoredPath    string       `gorm:"uniqueIndex;not null"` // Caminho completo onde a foto está armazenada
	ThumbnailPath string       // Caminho para a miniatura (opcional, para futuras implementações)
	UploadDate    time.Time    // Data/hora do upload
	ExifDate      *time.Time   // Data/hora da foto extraída do EXIF (pode ser nula)
	Hash          string       `gorm:"uniqueIndex;not null"` // Hash da foto para detecção de duplicatas
	FileSize      int64        // Tamanho do arquivo em bytes
	MimeType      string       // Tipo MIME do arquivo (ex: image/jpeg)
	Width         int          // Largura da imagem em pixels
	Height        int          // Altura da imagem em pixels
	Description   string       // Descrição ou legenda da foto
	Tags          string       // Tags da foto, armazenadas como string separada por vírgulas (ex: "viagem,praia")
	AlbumPhotos   []AlbumPhoto // Relação com a tabela de junção AlbumPhoto
}

// Album representa um álbum personalizado de fotos.
type Album struct {
	gorm.Model
	Name        string       `gorm:"uniqueIndex;not null"` // Nome do álbum
	Description string       // Descrição do álbum
	AlbumPhotos []AlbumPhoto // Relação com a tabela de junção AlbumPhoto
}

// AlbumPhoto é uma tabela de junção para a relação muitos-para-muitos entre Photo e Album.
type AlbumPhoto struct {
	gorm.Model
	PhotoID uint  // ID da foto
	Photo   Photo `gorm:"foreignkey:PhotoID"`
	AlbumID uint  // ID do álbum
	Album   Album `gorm:"foreignkey:AlbumID"`
}
