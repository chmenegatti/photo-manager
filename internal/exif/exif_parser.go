package exif

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// ExifData contém os metadados EXIF relevantes para a foto.
type ExifData struct {
	DateTime *time.Time // Data e hora da criação da foto
	// Outros campos EXIF podem ser adicionados aqui conforme necessidade (e.g., Make, Model, GPS)
}

// ExtractExifData extrai metadados EXIF de um arquivo de imagem.
func ExtractExifData(filePath string) (*ExifData, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("não foi possível abrir o arquivo para leitura EXIF: %w", err)
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		// É comum que imagens não tenham dados EXIF, não tratamos isso como erro fatal.
		// Apenas retornamos nil para ExifData e nil para erro, indicando que não há dados EXIF.
		if err == io.EOF || err.Error() == "no exif data" { // goexif retorna io.EOF para arquivos sem EXIF
			return nil, nil
		}
		return nil, fmt.Errorf("não foi possível decodificar dados EXIF: %w", err)
	}

	var exifData ExifData
	// Tenta extrair a data e hora de criação.
	tm, err := x.DateTime()
	if err == nil {
		exifData.DateTime = &tm
	} else {
		// Logar o erro se a data não puder ser extraída, mas não falhar
		fmt.Printf("Aviso: Não foi possível extrair DateTime EXIF: %v\n", err)
	}

	// Adicione aqui a extração de outros campos EXIF se necessário
	// Exemplo:
	// camModel, err := x.Get(exif.Model)
	// if err == nil {
	// 	if modelStr, err := camModel.String(); err == nil {
	// 		fmt.Println("Modelo da Câmera:", modelStr)
	// 	}
	// }

	if exifData.DateTime == nil {
		return nil, nil // Não há dados EXIF relevantes para retornar
	}

	return &exifData, nil
}
