// Package tasks_service содержит бизнес-логику управления задачами.
// Сервис оркестрирует работу с репозиторием: валидирует входные данные,
// создаёт доменные объекты и применяет бизнес-правила перед сохранением.
package tasks_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/milas1221/golang-todoapp/internal/core/domain"
)

// TasksService — сервис задач. Содержит бизнес-логику, не завязанную
// на детали хранилища или транспортного протокола.
type TasksService struct {
	tasksRepository TasksRepository
}

// TasksRepository — интерфейс репозитория задач, который должна реализовать
// конкретная реализация хранилища (например, PostgreSQL).
//
// Определение интерфейса в пакете сервиса (а не репозитория) — это паттерн
// «Interface Segregation» и «Dependency Inversion»: сервис владеет контрактом,
// а репозиторий его выполняет. Это упрощает тестирование через mock-объекты.
type TasksRepository interface {
	SaveTask(
		ctx context.Context,
		task domain.Task,
	) (domain.Task, error)

	GetTasks(
		ctx context.Context,
		userID *uuid.UUID,
		limit *int,
		offset *int,
	) ([]domain.Task, error)

	GetTask(
		ctx context.Context,
		id uuid.UUID,
	) (domain.Task, error)

	DeleteTask(
		ctx context.Context,
		id uuid.UUID,
	) error

	UpdateTask(
		ctx context.Context,
		task domain.Task,
	) (domain.Task, error)
}

// NewTasksService создаёт сервис задач с внедрённым репозиторием.
func NewTasksService(
	tasksRepository TasksRepository,
) *TasksService {
	return &TasksService{
		tasksRepository: tasksRepository,
	}
}
