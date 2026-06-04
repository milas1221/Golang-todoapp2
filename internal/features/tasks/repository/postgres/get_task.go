package tasks_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/milas1221/golang-todoapp/internal/core/domain"
	core_errors "github.com/milas1221/golang-todoapp/internal/core/errors"
	core_postgres_pool "github.com/milas1221/golang-todoapp/internal/core/repository/postgres/pool"
)

// GetTask возвращает задачу по ID.
// Если задача не найдена — ErrNoRows из пула транслируется в core_errors.ErrNotFound,
// что приведёт к HTTP 404 в транспортном слое.
func (r *TasksRepository) GetTask(
	ctx context.Context,
	id uuid.UUID,
) (domain.Task, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, version, title, description, completed, created_at, completed_at, author_user_id
	FROM todoapp.tasks
	WHERE id=$1;
	`

	row := r.pool.QueryRow(ctx, query, id)

	var taskModel TaskModel
	if err := taskModel.Scan(row); err != nil {
		if errors.Is(err, core_postgres_pool.ErrNoRows) {
			return domain.Task{}, fmt.Errorf(
				"task with id='%s': %w",
				id,
				core_errors.ErrNotFound,
			)
		}

		return domain.Task{}, fmt.Errorf("scan error: %w", err)
	}

	taskDomain := modelToDomain(taskModel)

	return taskDomain, nil
}
