// Package web_service содержит сервис для отдачи статических веб-страниц.
package web_service

import "github.com/milas1221/golang-todoapp/internal/core/domain"

// WebService — сервис для работы с веб-страницами приложения.
type WebService struct {
	webRepository WebRepository
}

// WebRepository — интерфейс репозитория для чтения файлов.
type WebRepository interface {
	GetFile(filePath string) (domain.File, error)
}

// NewWebService создаёт сервис с внедрённым репозиторием файловой системы.
func NewWebService(
	webRepository WebRepository,
) *WebService {
	return &WebService{
		webRepository: webRepository,
	}
}
