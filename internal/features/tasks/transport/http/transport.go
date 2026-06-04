// Package tasks_transport_http содержит HTTP-обработчики для фичи задач.
// Каждый обработчик: читает запрос → вызывает сервис → формирует ответ.
// Транспортный слой не содержит бизнес-логики — только преобразование данных.
package tasks_transport_http

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/milas1221/golang-todoapp/internal/core/domain"
	core_http_server "github.com/milas1221/golang-todoapp/internal/core/transport/http/server"
)

// TasksHTTPHandler — HTTP-обработчик для операций с задачами.
type TasksHTTPHandler struct {
	tasksService TasksService
}

// TasksService — интерфейс сервиса задач, который должен быть реализован
// для работы обработчика. Определён здесь по принципу DIP:
// транспортный слой владеет контрактом зависимости.
type TasksService interface {
	CreateTask(
		ctx context.Context,
		title string,
		description *string,
		authorUserID uuid.UUID,
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

	PatchTask(
		ctx context.Context,
		id uuid.UUID,
		patch domain.TaskPatch,
	) (domain.Task, error)
}

// NewTasksHTTPHandler создаёт обработчик задач с внедрённым сервисом.
func NewTasksHTTPHandler(
	tasksService TasksService,
) *TasksHTTPHandler {
	return &TasksHTTPHandler{
		tasksService: tasksService,
	}
}

// Routes возвращает маршруты REST API для задач.
// Вызывается в main.go при регистрации маршрутов в APIVersionRouter.
func (h *TasksHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/tasks",
			Handler: h.CreateTask,
		},
		{
			Method:  http.MethodGet,
			Path:    "/tasks",
			Handler: h.GetTasks,
		},
		{
			Method:  http.MethodGet,
			Path:    "/tasks/{id}",
			Handler: h.GetTask,
		},
		{
			Method:  http.MethodDelete,
			Path:    "/tasks/{id}",
			Handler: h.DeleteTask,
		},
		{
			Method:  http.MethodPatch,
			Path:    "/tasks/{id}",
			Handler: h.PatchTask,
		},
	}
}
