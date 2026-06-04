// Package web_fs_repository реализует репозиторий для чтения статических файлов
// из файловой системы. Использует стандартный os.ReadFile.
package web_fs_repository

// WebRepository — репозиторий для доступа к файлам веб-интерфейса.
// Пустая структура: нет состояния, только методы.
type WebRepository struct{}

// NewWebRepository создаёт репозиторий для файловой системы.
func NewWebRepository() *WebRepository {
	return &WebRepository{}
}
