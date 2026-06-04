package tasks_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/milas1221/golang-todoapp/internal/core/domain"
)

// PatchTask частично обновляет задачу по ID.
//
// Паттерн «read-modify-write» с оптимистичной блокировкой:
//  1. GetTask — читаем текущее состояние задачи (включая актуальный version)
//  2. ApplyPatch — применяем изменения к копии (с внутренней валидацией)
//  3. UpdateTask — сохраняем; WHERE version=$N защищает от конкурентного изменения
func (s *TasksService) PatchTask(
	ctx context.Context,
	id uuid.UUID,
	patch domain.TaskPatch,
) (domain.Task, error) {
	task, err := s.tasksRepository.GetTask(ctx, id)
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task from repository: %w", err)
	}

	if err := task.ApplyPatch(patch); err != nil {
		return domain.Task{}, fmt.Errorf("apply task patch: %w", err)
	}

	patchedTask, err := s.tasksRepository.UpdateTask(ctx, task)
	if err != nil {
		return domain.Task{}, fmt.Errorf("update task in repository: %w", err)
	}

	return patchedTask, nil
}
