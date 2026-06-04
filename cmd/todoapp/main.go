// Точка входа приложения. Здесь происходит:
//   - Инициализация конфигурации и логгера
//   - Подключение к базе данных PostgreSQL
//   - «Сборка» всех фич (Repository → Service → HTTP Handler)
//   - Запуск HTTP-сервера с graceful shutdown
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	core_config "github.com/milas1221/golang-todoapp/internal/core/config"
	core_logger "github.com/milas1221/golang-todoapp/internal/core/logger"
	core_pgx_pool "github.com/milas1221/golang-todoapp/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "github.com/milas1221/golang-todoapp/internal/core/transport/http/middleware"
	core_http_server "github.com/milas1221/golang-todoapp/internal/core/transport/http/server"
	statistics_postgres_repository "github.com/milas1221/golang-todoapp/internal/features/statistics/repository/postgres"
	statistics_service "github.com/milas1221/golang-todoapp/internal/features/statistics/service"
	statistics_transport_http "github.com/milas1221/golang-todoapp/internal/features/statistics/transport/http"
	tasks_postgres_repository "github.com/milas1221/golang-todoapp/internal/features/tasks/repository/postgres"
	tasks_service "github.com/milas1221/golang-todoapp/internal/features/tasks/service"
	tasks_transport_http "github.com/milas1221/golang-todoapp/internal/features/tasks/transport/http"
	users_postgres_repository "github.com/milas1221/golang-todoapp/internal/features/users/repository/postgres"
	users_service "github.com/milas1221/golang-todoapp/internal/features/users/service"
	users_transport_http "github.com/milas1221/golang-todoapp/internal/features/users/transport/http"
	web_fs_repository "github.com/milas1221/golang-todoapp/internal/features/web/repository/file_system"
	web_service "github.com/milas1221/golang-todoapp/internal/features/web/service"
	web_transport_http "github.com/milas1221/golang-todoapp/internal/features/web/transport/http"
	"go.uber.org/zap"

	// Импорт пакета docs регистрирует Swagger-спецификацию в глобальной переменной.
	// Сам пакет не используется напрямую — нужен только side-effect от init().
	_ "github.com/milas1221/golang-todoapp/docs"
)

// Аннотации для автогенерации Swagger-документации (swaggo/swag).
// @title        Golang Todo API
// @version      1.0
// @description  Todo Application REST-API scheme
// @host         127.0.0.1:5050
// @BasePath     /api/v1
func main() {
	// Загружаем общую конфигурацию приложения
	// NewConfigMust — паттерн «Must»: паникует при ошибке, т.к. на старте
	// приложение не может продолжать работу с невалидной конфигурацией.
	cfg := core_config.NewConfigMust()
	time.Local = cfg.TimeZone

	// Создаём корневой контекст, который отменяется при получении SIGINT/SIGTERM
	// (Ctrl+C или команда `kill`). Это основа для graceful shutdown.
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	// Инициализируем логгер приложения
	// Пишет одновременно в stdout и в файл (см. internal/core/logger).
	logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("failed to init application logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("application time zone", zap.Any("zone", time.Local))

	// Создаём пулл соединений с PostgreSQL через библиотеку pgx.
	// Пул переиспользует соединения, что гораздо эффективнее,
	// чем открывать новое соединение на каждый SQL запрос.
	logger.Debug("initializing postgres connection pool")
	postgresPool, err := core_pgx_pool.NewPool(
		ctx,
		core_pgx_pool.NewConfigMust(),
	)
	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}
	defer postgresPool.Close()

	// Ручное внедрение зависимостей (Dependency Injection):
	// Repository → Service → HTTP Handler.
	// Каждый слой знает только об интерфейсе нижележащего.
	// Это обеспечивает слабую связанность (loose coupling) и тестируемость.

	logger.Debug("initializing feature", zap.String("feature", "users"))
	usersRepository := users_postgres_repository.NewUsersRepository(postgresPool)
	usersService := users_service.NewUsersService(usersRepository)
	usersTransportHTTP := users_transport_http.NewUsersHTTPHandler(usersService)

	logger.Debug("initializing feature", zap.String("feature", "tasks"))
	tasksRepository := tasks_postgres_repository.NewTasksRepository(postgresPool)
	tasksService := tasks_service.NewTasksService(tasksRepository)
	tasksTransportHTTP := tasks_transport_http.NewTasksHTTPHandler(tasksService)

	logger.Debug("initializing feature", zap.String("feature", "statistics"))
	statisticsRepository := statistics_postgres_repository.NewStatisticsRepository(postgresPool)
	statisticsService := statistics_service.NewStatisticsService(statisticsRepository)
	statisticsTransportHTTP := statistics_transport_http.NewStatisticsHTTPHandler(statisticsService)

	logger.Debug("initializing feature", zap.String("feature", "web"))
	webRepository := web_fs_repository.NewWebRepository()
	webService := web_service.NewWebService(webRepository)
	webTransportHTTP := web_transport_http.NewWebHTTPHandler(webService)

	// Собираем HTTP-сервер с цепочкой middleware.
	// Middleware применяются ко всем маршрутам (Route) в порядке объявления:
	// CORS → RequestID → Logger → Trace → Panic recovery.
	logger.Debug("initializing HTTP server")
	httpConfig := core_http_server.NewConfigMust()
	httpServer := core_http_server.NewHTTPServer(
		httpConfig,
		logger,
		core_http_middleware.CORS(httpConfig.AllowedOrigins),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)

	// Регистрируем маршруты API v1.
	// APIVersionRouter автоматически добавляет префикс /api/v1 ко всем путям.
	apiVersionRouterV1 := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
	apiVersionRouterV1.AddRoutes(usersTransportHTTP.Routes()...)
	apiVersionRouterV1.AddRoutes(tasksTransportHTTP.Routes()...)
	apiVersionRouterV1.AddRoutes(statisticsTransportHTTP.Routes()...)

	/*
		Пример регистрации API v2 с отдельными middleware:

		apiVersionRouterV2 := core_http_server.NewAPIVersionRouter(
			core_http_server.ApiVersion2,
			core_http_middleware.Dummy("api v2 middleware"),
		)
		apiVersionRouterV2.RegisterRoutes(usersTransportHTTP.Routes()...)
	*/

	httpServer.RegisterAPIRouters(
		apiVersionRouterV1,
		// apiVersionRouterV2,
	)
	// Регистрируем маршруты, не принадлежащие к какой-то определённой версии API (главная страница).
	httpServer.RegisterRoutes(webTransportHTTP.Routes()...)
	// Регистрируем Swagger UI по адресу /swagger/.
	httpServer.RegisterSwagger()

	// Запускаем сервер. Блокируется до получения сигнала завершения.
	// После сигнала выполняет graceful shutdown: ждёт завершения активных запросов.
	if err := httpServer.Run(ctx); err != nil {
		logger.Error("HTTP server run error", zap.Error(err))
	}
}
