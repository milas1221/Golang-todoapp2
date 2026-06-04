package tasks_postgres_repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/milas1221/golang-todoapp/internal/core/domain"
	core_postgres_pool "github.com/milas1221/golang-todoapp/internal/core/repository/postgres/pool"
)

// TaskModel — структура для маппинга строки таблицы `todoapp.tasks` в Go-тип.
// Порядок полей совпадает с порядком столбцов в SELECT-запросах репозитория —
// именно в этом порядке pgx заполняет поля при Scan.
type TaskModel struct {
	ID           uuid.UUID
	Version      int
	Title        string
	Description  *string
	Completed    bool
	CreatedAt    time.Time
	CompletedAt  *time.Time
	AuthorUserID uuid.UUID
}

// Scan заполняет поля модели из результата запроса к БД.
// Принимает интерфейс Row (не pgx.Row напрямую) для изоляции от драйвера.
func (m *TaskModel) Scan(row core_postgres_pool.Row) error {
	return row.Scan(
		&m.ID,
		&m.Version,
		&m.Title,
		&m.Description,
		&m.Completed,
		&m.CreatedAt,
		&m.CompletedAt,
		&m.AuthorUserID,
	)
}

// modelToDomain конвертирует модель БД в доменный объект.
// Это «антикоррупционный слой» между хранилищем и бизнес-логикой:
// сервис работает только с domain.Task, не зная о структуре БД.
func modelToDomain(model TaskModel) domain.Task {
	return domain.NewTask(
		model.ID,
		model.Version,
		model.Title,
		model.Description,
		model.Completed,
		model.CreatedAt,
		model.CompletedAt,
		model.AuthorUserID,
	)
}

// modelsToDomains конвертирует список моделей БД в список доменных объектов.
func modelsToDomains(models []TaskModel) []domain.Task {
	domains := make([]domain.Task, len(models))

	for i, model := range models {
		domains[i] = modelToDomain(model)
	}

	return domains
}
