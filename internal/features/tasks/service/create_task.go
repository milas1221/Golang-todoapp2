package tasks_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/milas1221/golang-todoapp/internal/core/domain"
)

// CreateTask создаёт новую задачу: формирует доменный объект, валидирует его
// и сохраняет через репозиторий.
//
// Порядок действий:
//  1. domain.CreateTask — генерирует UUID, version=1, completed=false, createdAt=now
//  2. task.Validate()   — проверяет инварианты (длина Title, корректность CompletedAt и т.д.)
//  3. repository.SaveTask — вставляет запись в БД и возвращает сохранённое состояние
func (s *TasksService) CreateTask(
	ctx context.Context,
	title string,
	description *string,
	authorUserID uuid.UUID,
) (domain.Task, error) {
	task := domain.CreateTask(
		title,
		description,
		authorUserID,
	)

	if err := task.Validate(); err != nil {
		return domain.Task{}, fmt.Errorf("validate task domain: %w", err)
	}

	task, err := s.tasksRepository.SaveTask(ctx, task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("save task in repository: %w", err)
	}

	return task, nil
}
