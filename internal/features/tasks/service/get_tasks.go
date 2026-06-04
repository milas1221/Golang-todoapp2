package tasks_service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/milas1221/golang-todoapp/internal/core/domain"
	core_errors "github.com/milas1221/golang-todoapp/internal/core/errors"
)

// GetTasks возвращает список задач с пагинацией и опциональным фильтром по автору.
// Валидирует limit и offset до обращения к репозиторию, чтобы не делать
// бессмысленный запрос с некорректными параметрами.
func (s *TasksService) GetTasks(
	ctx context.Context,
	userID *uuid.UUID,
	limit *int,
	offset *int,
) ([]domain.Task, error) {
	if limit != nil && *limit < 0 {
		return nil, fmt.Errorf(
			"limit must be non-negative: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	if offset != nil && *offset < 0 {
		return nil, fmt.Errorf(
			"offset must be non-negative: %w",
			core_errors.ErrInvalidArgument,
		)
	}

	tasks, err := s.tasksRepository.GetTasks(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get tasks from repository: %w", err)
	}

	return tasks, nil
}
