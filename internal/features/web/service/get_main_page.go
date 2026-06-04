package web_service

import (
	"fmt"
	"os"
	"path"

	"github.com/milas1221/golang-todoapp/internal/core/domain"
)

// GetMainPage возвращает содержимое главной HTML-страницы.
// Путь к файлу формируется относительно PROJECT_ROOT (переменная окружения),
// чтобы приложение работало корректно независимо от рабочей директории запуска.
func (s *WebService) GetMainPage() (domain.File, error) {
	htmlFilePath := path.Join(
		os.Getenv("PROJECT_ROOT"),
		"/public/index.html",
	)

	htmlFile, err := s.webRepository.GetFile(htmlFilePath)
	if err != nil {
		return domain.File{}, fmt.Errorf("get file from repository: %w", err)
	}

	return htmlFile, nil
}
