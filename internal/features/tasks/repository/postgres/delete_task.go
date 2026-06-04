package tasks_postgres_repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	core_errors "github.com/milas1221/golang-todoapp/internal/core/errors"
)

// DeleteTask удаляет задачу по ID.
// Проверяем cmdTag.RowsAffected() == 0 вместо ErrNoRows, т.к. DELETE
// не возвращает строки — используем количество удалённых записей для
// определения факта существования задачи.
func (r *TasksRepository) DeleteTask(
	ctx context.Context,
	id uuid.UUID,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	DELETE FROM todoapp.tasks
	WHERE id=$1;
	`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("exec query: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf(
			"task with id='%s': %w",
			id,
			core_errors.ErrNotFound,
		)
	}

	return nil
}
