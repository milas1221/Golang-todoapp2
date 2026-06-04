package web_fs_repository

import (
	"errors"
	"fmt"
	"os"

	"github.com/milas1221/golang-todoapp/internal/core/domain"
	core_errors "github.com/milas1221/golang-todoapp/internal/core/errors"
)

// GetFile читает файл по пути filePath и возвращает его содержимое как domain.File.
// Преобразует os.ErrNotExist в core_errors.ErrNotFound для единообразной обработки
// на уровне транспортного слоя (→ HTTP 404).
func (r *WebRepository) GetFile(filePath string) (domain.File, error) {
	buffer, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.File{}, fmt.Errorf(
				"file: %s: %w",
				filePath,
				core_errors.ErrNotFound,
			)
		}

		return domain.File{}, fmt.Errorf(
			"get file: %s: %w",
			filePath,
			err,
		)
	}

	file := domain.NewFile(buffer)

	return file, nil
}
