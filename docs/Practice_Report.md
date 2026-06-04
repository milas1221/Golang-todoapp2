МИНИСТЕРСТВО НАУКИ И ВЫСШЕГО ОБРАЗОВАНИЯ РОССИЙСКОЙ ФЕДЕРАЦИИ

ФГБОУ ВО «ДАГЕСТАНСКИЙ ГОСУДАРСТВЕННЫЙ ТЕХНИЧЕСКИЙ УНИВЕРСИТЕТ»

Факультет информатики и вычислительной техники
Кафедра информатики

ОТЧЁТ
по производственной практике

Тема:
«Разработка RESTful веб-приложения для управления задачами с использованием языка программирования Go и многослойной архитектуры»

Выполнил: студент группы [___]
Берсиров С.

Руководитель: [___]

Махачкала 2026

---

ИНДИВИДУАЛЬНОЕ ЗАДАНИЕ

Разработать RESTful веб-приложение TodoApp для управления задачами на языке программирования Go с использованием многослойной архитектуры, СУБД PostgreSQL, контейнеризации Docker и автоматической документации Swagger.

В ходе производственной практики необходимо:
— реализовать серверную часть приложения на языке Go с применением многослойной архитектуры (Transport — Service — Repository);
— реализовать модули управления пользователями (CRUD), задачами (CRUD) и статистики;
— обеспечить хранение данных в СУБД PostgreSQL с управлением схемой через миграции;
— реализовать REST API с автоматической Swagger-документацией;
— обеспечить контейнеризацию компонентов системы с помощью Docker Compose;
— провести функциональное тестирование реализованных эндпоинтов;
— подготовить веб-интерфейс для работы с задачами.

1.1 Требования к разрабатываемому программному обеспечению

Функциональные требования:
— создание, получение, обновление и удаление пользователей;
— создание, получение, обновление и удаление задач;
— привязка задач к пользователям через поле author_user_id;
— автоматическая фиксация времени завершения задачи при изменении статуса;
— получение статистики по задачам с фильтрацией по пользователю и временному диапазону;
— пагинация списков (параметры limit и offset);
— валидация входных данных (длина полей, формат телефона, обязательные поля);
— интерактивная Swagger-документация по адресу /swagger/;
— веб-интерфейс по адресу /.

Нефункциональные требования:
— graceful shutdown с таймаутом 30 секунд;
— пул соединений с PostgreSQL (библиотека pgx);
— структурированное логирование (библиотека zap) в файл и stdout;
— поддержка CORS;
— оптимистичная блокировка при обновлении записей;
— контейнеризация через Docker Compose.

---

СОДЕРЖАНИЕ

ИНДИВИДУАЛЬНОЕ ЗАДАНИЕ ...................................... 2
1.1 Требования к разрабатываемому ПО ....................... 2
ВВЕДЕНИЕ ................................................... 4
1 Алгоритм работы приложения ............................... 5
  1.1 Общий алгоритм обработки HTTP-запросов ............... 5
  1.2 Алгоритм создания задачи ............................ 6
  1.3 Алгоритм обновления задачи (PATCH) .................. 7
2 Модульная структура приложения ........................... 8
  2.1 Модуль core (ядро) .................................. 8
  2.2 Модуль users (пользователи) ......................... 9
  2.3 Модуль tasks (задачи) ............................... 9
  2.4 Модуль statistics (статистика) ....................... 9
  2.5 Модуль web (веб-интерфейс) .......................... 10
3 Реализация ............................................... 11
  3.1 Доменный слой ....................................... 11
  3.2 Сервисный слой ...................................... 13
  3.3 Слой репозитория .................................... 14
  3.4 Транспортный слой (HTTP) ............................ 16
  3.5 Конфигурация и запуск ............................... 18
4 Тестирование ............................................. 20
5 Оценка результатов ....................................... 22
ЗАКЛЮЧЕНИЕ ................................................. 23
СПИСОК ИСПОЛЬЗОВАННЫХ ИСТОЧНИКОВ ........................... 24
ПРИЛОЖЕНИЕ А. Исходный код main.go ......................... 25
ПРИЛОЖЕНИЕ Б. Исходный код доменных моделей ................ 27
ПРИЛОЖЕНИЕ В. Файл docker-compose.yaml ..................... 30

---

ВВЕДЕНИЕ

Производственная практика является важным этапом подготовки специалиста в области информационных технологий, позволяющим закрепить теоретические знания путём решения практических задач разработки программного обеспечения.

Целью производственной практики является разработка RESTful веб-приложения для управления задачами (TodoApp) на языке программирования Go с применением многослойной архитектуры, СУБД PostgreSQL и технологий контейнеризации Docker.

Задачи практики:
— реализовать серверную часть приложения с разделением на слои Transport, Service и Repository;
— реализовать модули управления пользователями, задачами и статистикой;
— обеспечить хранение данных в PostgreSQL с управлением схемой через миграции;
— реализовать REST API с Swagger-документацией;
— провести функциональное тестирование;
— подготовить документацию.

Практика выполнена на базе проекта TodoApp — веб-приложения с открытым исходным кодом, размещённого в репозитории GitHub. Приложение реализует полный цикл CRUD-операций для пользователей и задач, предоставляет REST API с автоматической документацией и веб-интерфейс для визуального управления задачами.

---

1 Алгоритм работы приложения

1.1 Общий алгоритм обработки HTTP-запросов

Обработка каждого HTTP-запроса в приложении проходит через цепочку компонентов, организованных в конвейер (pipeline). Алгоритм обработки включает следующие этапы.

Рисунок 1 — Блок-схема обработки HTTP-запроса

[HTTP-запрос от клиента]
        ↓
[Middleware: CORS] — проверка заголовков Origin
        ↓
[Middleware: RequestID] — генерация уникального идентификатора запроса
        ↓
[Middleware: Logger] — логирование метода, пути, статуса и времени обработки
        ↓
[Middleware: Trace] — добавление трассировочной информации в контекст
        ↓
[Middleware: Panic] — перехват паник, формирование HTTP 500
        ↓
[Router (ServeMux)] — маршрутизация по методу и пути
        ↓
[Handler] — десериализация запроса, валидация, вызов сервиса
        ↓
[Service] — бизнес-логика, создание/модификация доменных объектов
        ↓
[Repository] — выполнение SQL-запросов к PostgreSQL
        ↓
[Handler] — сериализация ответа в JSON
        ↓
[HTTP-ответ клиенту]

Middleware применяются ко всем маршрутам сервера и выполняются в порядке их регистрации. Каждый middleware оборачивает следующий обработчик, формируя цепочку вызовов. Данный подход реализован через функцию ChainMiddleware, которая последовательно применяет middleware к корневому обработчику (ServeMux).

1.2 Алгоритм создания задачи

Алгоритм создания задачи включает следующие шаги:
1) клиент отправляет POST-запрос на /api/v1/tasks с телом в формате JSON, содержащим поля title, description (опционально) и author_user_id;
2) транспортный слой десериализует тело запроса и выполняет валидацию через библиотеку go-playground/validator;
3) сервисный слой вызывает конструктор domain.CreateTask, который автоматически генерирует UUID, устанавливает version = 1, completed = false, created_at = now();
4) сервис вызывает метод Validate() доменной сущности для проверки инвариантов (длина заголовка, описания);
5) репозиторий выполняет SQL-запрос INSERT INTO tasks ... RETURNING для сохранения и получения записи из БД;
6) транспортный слой формирует JSON-ответ с HTTP-статусом 201 Created.

1.3 Алгоритм обновления задачи (PATCH с оптимистичной блокировкой)

Обновление задачи реализовано через механизм PATCH-запросов с оптимистичной блокировкой. Алгоритм:
1) клиент отправляет PATCH-запрос с частичными данными (только изменяемые поля);
2) транспортный слой десериализует запрос с использованием типа Nullable[T] для различения «не передано», «null» и «значение»;
3) сервис получает текущую задачу из БД (GetTask);
4) к задаче применяется метод ApplyPatch, который работает с копией объекта, применяет изменения и проверяет валидность результата;
5) при изменении статуса completed на true автоматически устанавливается completed_at = now(), при изменении на false — completed_at сбрасывается в nil;
6) репозиторий выполняет UPDATE tasks SET ... WHERE id = $1 AND version = $2, включая условие по версии;
7) если версия изменилась (другой запрос успел обновить запись), UPDATE не затронет строк и система вернёт HTTP 409 Conflict;
8) при успешном обновлении version инкрементируется.

---

2 Модульная структура приложения

Приложение организовано по принципу feature-based structure, при котором каждый функциональный модуль содержит полный набор слоёв (transport, service, repository) и может быть разработан и протестирован независимо.

Рисунок 2 — Модульная диаграмма приложения

                    ┌──────────────────────┐
                    │      cmd/todoapp     │
                    │      (main.go)       │
                    └──────────┬───────────┘
                               │
            ┌──────────────────┼──────────────────┐
            │                  │                  │
    ┌───────┴───────┐  ┌──────┴──────┐  ┌───────┴───────┐
    │    features/  │  │  features/  │  │  features/    │
    │    users      │  │  tasks      │  │  statistics   │
    │  ┌──────────┐ │  │ ┌─────────┐│  │ ┌───────────┐ │
    │  │transport │ │  │ │transport││  │ │ transport  │ │
    │  │service   │ │  │ │service  ││  │ │ service    │ │
    │  │repository│ │  │ │reposit. ││  │ │ repository │ │
    │  └──────────┘ │  │ └─────────┘│  │ └───────────┘ │
    └───────────────┘  └────────────┘  └───────────────┘
            │                  │                  │
            └──────────────────┼──────────────────┘
                               │
                    ┌──────────┴───────────┐
                    │    internal/core     │
                    │  domain, server,     │
                    │  middleware, logger,  │
                    │  pool, config        │
                    └──────────────────────┘

2.1 Модуль core (ядро)

Модуль core содержит общие компоненты, используемые всеми функциональными модулями:
— domain — доменные сущности Task, User и тип Nullable[T];
— errors — общие типы ошибок (ErrNotFound, ErrInvalidArgument, ErrConflict);
— config — загрузка конфигурации из переменных окружения через библиотеку envconfig;
— logger — обёртка над zap-логгером с записью в файл и stdout;
— repository/postgres/pool — абстракция пула соединений с PostgreSQL;
— transport/http/server — HTTP-сервер с поддержкой версионирования API;
— transport/http/middleware — цепочка middleware (CORS, RequestID, Logger, Trace, Panic);
— transport/http/request — десериализация и валидация HTTP-запросов;
— transport/http/response — формирование HTTP-ответов и обработка ошибок;
— transport/http/types — тип NullableJSON для десериализации PATCH-запросов.

2.2 Модуль users (пользователи)

Модуль users реализует CRUD-операции с пользователями:
— transport/http — обработчики CreateUser, GetUsers, GetUser, PatchUser, DeleteUser;
— service — бизнес-логика: создание доменного объекта, валидация, применение патча;
— repository/postgres — SQL-запросы: INSERT, SELECT, UPDATE, DELETE с оптимистичной блокировкой.

Интерфейс UsersRepository определён в пакете сервиса и включает методы: SaveUser, GetUsers, GetUser, DeleteUser, UpdateUser.

2.3 Модуль tasks (задачи)

Модуль tasks реализует CRUD-операции с задачами:
— transport/http — обработчики CreateTask, GetTasks, GetTask, PatchTask, DeleteTask;
— service — бизнес-логика: создание задачи с автоматическими полями, применение патча с автоматическим управлением completed_at;
— repository/postgres — SQL-запросы с поддержкой фильтрации по user_id и пагинации.

Интерфейс TasksRepository определён в пакете сервиса и включает методы: SaveTask, GetTasks, GetTask, DeleteTask, UpdateTask.

2.4 Модуль statistics (статистика)

Модуль statistics реализует получение статистики по задачам:
— transport/http — обработчик GetStatistics с параметрами фильтрации (user_id, from, to);
— service — получение задач из репозитория и расчёт статистики;
— repository/postgres — SQL-запрос с фильтрацией по пользователю и временному диапазону.

Интерфейс StatisticsRepository требует только метод GetTasks — пример принципа Interface Segregation (ISP из SOLID).

2.5 Модуль web (веб-интерфейс)

Модуль web обеспечивает обслуживание веб-интерфейса приложения:
— transport/http — обработчик GetMainPage, обслуживающий HTML-файл;
— service — бизнес-логика получения файла;
— repository/file_system — чтение файлов из файловой системы.

---

3 Реализация

3.1 Доменный слой

Доменный слой содержит сущности и бизнес-правила, не зависящие от деталей реализации (БД, HTTP). Основные сущности: Task и User.

Структура сущности Task:

    type Task struct {
        ID           uuid.UUID
        Version      int
        Title        string
        Description  *string
        Completed    bool
        CreatedAt    time.Time
        CompletedAt  *time.Time
        AuthorUserID uuid.UUID
    }

Конструктор CreateTask генерирует UUID, устанавливает начальную версию и текущее время:

    func CreateTask(
        title string,
        description *string,
        authorUserID uuid.UUID,
    ) Task {
        var (
            id                     = uuid.New()
            version                = 1
            completed              = false
            createdAt              = time.Now()
            completedAt *time.Time = nil
        )
        return NewTask(id, version, title, description,
            completed, createdAt, completedAt, authorUserID)
    }

Метод Validate() проверяет инварианты доменной модели. Для корректного подсчёта длины Unicode-строк используется len([]rune(...)):

    func (t *Task) Validate() error {
        titleLen := len([]rune(t.Title))
        if titleLen < 1 || titleLen > 100 {
            return fmt.Errorf(
                "invalid `Title` len: %d: %w",
                titleLen, core_errors.ErrInvalidArgument,
            )
        }
        // ... проверка Description, согласованности
        // Completed и CompletedAt
        return nil
    }

Метод ApplyPatch работает с копией объекта для обеспечения транзакционности: либо все изменения применяются, либо ни одно:

    func (t *Task) ApplyPatch(patch TaskPatch) error {
        if err := patch.Validate(); err != nil {
            return fmt.Errorf("validate task patch: %w", err)
        }
        tmp := *t // работаем с копией
        if patch.Title.Set {
            tmp.Title = *patch.Title.Value
        }
        if patch.Completed.Set {
            tmp.Completed = *patch.Completed.Value
            if tmp.Completed {
                completedAt := time.Now()
                tmp.CompletedAt = &completedAt
            } else {
                tmp.CompletedAt = nil
            }
        }
        if err := tmp.Validate(); err != nil {
            return fmt.Errorf("validate patched task: %w", err)
        }
        *t = tmp // замена оригинала только после валидации
        return nil
    }

Структура сущности User:

    type User struct {
        ID          uuid.UUID
        Version     int
        FullName    string
        PhoneNumber *string
    }

Метод Validate() проверяет длину имени (3–100 символов) и формат телефона (начинается с «+», далее только цифры, длина 10–15 символов):

    func (u *User) Validate() error {
        fullNameLen := len([]rune(u.FullName))
        if fullNameLen < 3 || fullNameLen > 100 {
            return fmt.Errorf(
                "invalid `FullName` len: %d: %w",
                fullNameLen, core_errors.ErrInvalidArgument,
            )
        }
        if u.PhoneNumber != nil {
            re := regexp.MustCompile(`^\+[0-9]+$`)
            if !re.MatchString(*u.PhoneNumber) {
                return fmt.Errorf(
                    "invalid `PhoneNumber` format: %w",
                    core_errors.ErrInvalidArgument,
                )
            }
        }
        return nil
    }

3.2 Сервисный слой

Сервисный слой содержит бизнес-логику и оркестрирует взаимодействие между транспортным слоем и репозиторием. Интерфейс репозитория определён в пакете сервиса по принципу Dependency Inversion.

Интерфейс TasksRepository:

    type TasksRepository interface {
        SaveTask(ctx context.Context, task domain.Task) (domain.Task, error)
        GetTasks(ctx context.Context, userID *uuid.UUID,
            limit *int, offset *int) ([]domain.Task, error)
        GetTask(ctx context.Context, id uuid.UUID) (domain.Task, error)
        DeleteTask(ctx context.Context, id uuid.UUID) error
        UpdateTask(ctx context.Context, task domain.Task) (domain.Task, error)
    }

Создание сервиса с внедрённым репозиторием:

    type TasksService struct {
        tasksRepository TasksRepository
    }

    func NewTasksService(
        tasksRepository TasksRepository,
    ) *TasksService {
        return &TasksService{
            tasksRepository: tasksRepository,
        }
    }

Пример метода сервиса — создание задачи:

    func (s *TasksService) CreateTask(
        ctx context.Context,
        title string,
        description *string,
        authorUserID uuid.UUID,
    ) (domain.Task, error) {
        task := domain.CreateTask(title, description, authorUserID)
        if err := task.Validate(); err != nil {
            return domain.Task{}, fmt.Errorf("validate task: %w", err)
        }
        savedTask, err := s.tasksRepository.SaveTask(ctx, task)
        if err != nil {
            return domain.Task{}, fmt.Errorf("save task: %w", err)
        }
        return savedTask, nil
    }

Пример метода обновления задачи с оптимистичной блокировкой:

    func (s *TasksService) PatchTask(
        ctx context.Context,
        id uuid.UUID,
        patch domain.TaskPatch,
    ) (domain.Task, error) {
        task, err := s.tasksRepository.GetTask(ctx, id)
        if err != nil {
            return domain.Task{}, fmt.Errorf("get task: %w", err)
        }
        if err := task.ApplyPatch(patch); err != nil {
            return domain.Task{}, fmt.Errorf("apply patch: %w", err)
        }
        updatedTask, err := s.tasksRepository.UpdateTask(ctx, task)
        if err != nil {
            return domain.Task{}, fmt.Errorf("update task: %w", err)
        }
        return updatedTask, nil
    }

3.3 Слой репозитория

Слой репозитория инкапсулирует SQL-запросы к PostgreSQL. Взаимодействие с БД осуществляется через пул соединений pgx.

Создание репозитория:

    type TasksRepository struct {
        pool pool.PGXPool
    }

    func NewTasksRepository(pool pool.PGXPool) *TasksRepository {
        return &TasksRepository{pool: pool}
    }

Пример метода сохранения задачи (SaveTask):

    func (r *TasksRepository) SaveTask(
        ctx context.Context,
        task domain.Task,
    ) (domain.Task, error) {
        row := r.pool.QueryRow(
            ctx,
            `INSERT INTO tasks
                (id, version, title, description, completed,
                 created_at, completed_at, author_user_id)
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
             RETURNING id, version, title, description, completed,
                       created_at, completed_at, author_user_id`,
            task.ID, task.Version, task.Title, task.Description,
            task.Completed, task.CreatedAt, task.CompletedAt,
            task.AuthorUserID,
        )
        var model TaskModel
        err := row.Scan(
            &model.ID, &model.Version, &model.Title,
            &model.Description, &model.Completed,
            &model.CreatedAt, &model.CompletedAt,
            &model.AuthorUserID,
        )
        if err != nil {
            return domain.Task{}, fmt.Errorf("scan task: %w", err)
        }
        return model.ToDomain(), nil
    }

Пример метода обновления с оптимистичной блокировкой (UpdateTask):

    func (r *TasksRepository) UpdateTask(
        ctx context.Context,
        task domain.Task,
    ) (domain.Task, error) {
        row := r.pool.QueryRow(
            ctx,
            `UPDATE tasks
             SET version = version + 1,
                 title = $3, description = $4,
                 completed = $5, completed_at = $6
             WHERE id = $1 AND version = $2
             RETURNING id, version, title, description, completed,
                       created_at, completed_at, author_user_id`,
            task.ID, task.Version, task.Title, task.Description,
            task.Completed, task.CompletedAt,
        )
        // Если version изменилась, UPDATE не затронет строк →
        // Scan вернёт pgx.ErrNoRows → маппится в ErrConflict
        ...
    }

3.4 Транспортный слой (HTTP)

Транспортный слой обрабатывает HTTP-запросы, выполняет десериализацию и валидацию входных данных, вызывает сервисный слой и формирует HTTP-ответы.

Пример обработчика создания пользователя:

    type CreateUserRequest struct {
        FullName    string  `json:"full_name"
                             validate:"required,min=3,max=100"`
        PhoneNumber *string `json:"phone_number"
                             validate:"omitempty,min=10,max=15,
                             startswith=+"`
    }

    func (h *UsersHTTPHandler) CreateUser(
        rw http.ResponseWriter, r *http.Request,
    ) {
        ctx := r.Context()
        log := core_logger.FromContext(ctx)
        responseHandler :=
            core_http_response.NewHTTPResponseHandler(log, rw)

        var request CreateUserRequest
        if err := core_http_request.DecodeAndValidateRequest(
            r, &request,
        ); err != nil {
            responseHandler.ErrorResponse(err,
                "failed to decode and validate HTTP request")
            return
        }

        userDomain, err := h.usersService.CreateUser(
            ctx,
            request.FullName,
            request.PhoneNumber,
        )
        if err != nil {
            responseHandler.ErrorResponse(err,
                "failed to create user")
            return
        }

        response := CreateUserResponse(
            userDTOFromDomain(userDomain),
        )
        responseHandler.JSONResponse(response,
            http.StatusCreated)
    }

HTTP-сервер построен на стандартной библиотеке net/http (Go 1.22+) с поддержкой метода маршрутизации в шаблоне пути:

    type HTTPServer struct {
        mux        *http.ServeMux
        config     Config
        log        *core_logger.Logger
        middleware []core_http_middleware.Middleware
    }

Регистрация маршрутов с версионированием API:

    apiVersionRouterV1 := core_http_server.NewAPIVersionRouter(
        core_http_server.ApiVersion1,
    )
    apiVersionRouterV1.AddRoutes(usersTransportHTTP.Routes()...)
    apiVersionRouterV1.AddRoutes(tasksTransportHTTP.Routes()...)
    apiVersionRouterV1.AddRoutes(
        statisticsTransportHTTP.Routes()...,
    )

Graceful shutdown реализован через context.NotifyContext и server.Shutdown:

    func (s *HTTPServer) Run(ctx context.Context) error {
        mux := core_http_middleware.ChainMiddleware(
            s.mux, s.middleware...,
        )
        server := &http.Server{
            Addr:    s.config.Addr,
            Handler: mux,
        }
        ch := make(chan error, 1)
        go func() {
            defer close(ch)
            err := server.ListenAndServe()
            if !errors.Is(err, http.ErrServerClosed) {
                ch <- err
            }
        }()
        select {
        case err := <-ch:
            if err != nil {
                return fmt.Errorf("listen and serve: %w", err)
            }
        case <-ctx.Done():
            shutdownCtx, cancel := context.WithTimeout(
                context.Background(),
                s.config.ShutdownTimeout,
            )
            defer cancel()
            if err := server.Shutdown(shutdownCtx); err != nil {
                _ = server.Close()
                return fmt.Errorf("shutdown: %w", err)
            }
        }
        return nil
    }

3.5 Конфигурация и запуск

Точка входа приложения — функция main в файле cmd/todoapp/main.go. В ней выполняется ручное внедрение зависимостей: Repository → Service → HTTP Handler.

    func main() {
        cfg := core_config.NewConfigMust()
        time.Local = cfg.TimeZone

        ctx, cancel := signal.NotifyContext(
            context.Background(),
            syscall.SIGINT, syscall.SIGTERM,
        )
        defer cancel()

        logger, err := core_logger.NewLogger(
            core_logger.NewConfigMust(),
        )
        // ...

        postgresPool, err := core_pgx_pool.NewPool(
            ctx, core_pgx_pool.NewConfigMust(),
        )
        // ...

        // Внедрение зависимостей: Repository → Service → Handler
        usersRepository :=
            users_postgres_repository.NewUsersRepository(postgresPool)
        usersService :=
            users_service.NewUsersService(usersRepository)
        usersTransportHTTP :=
            users_transport_http.NewUsersHTTPHandler(usersService)

        // Аналогично для tasks, statistics, web...

        httpServer := core_http_server.NewHTTPServer(
            httpConfig, logger,
            core_http_middleware.CORS(httpConfig.AllowedOrigins),
            core_http_middleware.RequestID(),
            core_http_middleware.Logger(logger),
            core_http_middleware.Trace(),
            core_http_middleware.Panic(),
        )

        // Регистрация маршрутов
        apiVersionRouterV1 :=
            core_http_server.NewAPIVersionRouter(
                core_http_server.ApiVersion1,
            )
        apiVersionRouterV1.AddRoutes(
            usersTransportHTTP.Routes()...,
        )
        // ...
        httpServer.RegisterAPIRouters(apiVersionRouterV1)
        httpServer.RegisterSwagger()

        if err := httpServer.Run(ctx); err != nil {
            logger.Error("HTTP server error", zap.Error(err))
        }
    }

---

4 Тестирование

Функциональное тестирование проведено через Swagger UI и веб-интерфейс приложения. Результаты тестирования представлены в таблице 1.

Таблица 1 — Результаты функционального тестирования

| № | Тестовый сценарий                              | Ожидаемый результат          | Фактический результат        | Статус   |
|---|------------------------------------------------|------------------------------|------------------------------|----------|
| 1 | Создание пользователя с валидными данными      | HTTP 201, пользователь создан| HTTP 201, пользователь создан| Пройден  |
| 2 | Создание пользователя с невалидным телефоном   | HTTP 400, ошибка валидации   | HTTP 400, ошибка валидации   | Пройден  |
| 3 | Получение списка пользователей                 | HTTP 200, массив             | HTTP 200, массив             | Пройден  |
| 4 | Создание задачи с существующим автором         | HTTP 201, задача создана     | HTTP 201, задача создана     | Пройден  |
| 5 | Создание задачи без указания автора            | HTTP 400, ошибка валидации   | HTTP 400, ошибка валидации   | Пройден  |
| 6 | Обновление статуса задачи на completed = true  | HTTP 200, completed_at != null| HTTP 200, completed_at заполнено | Пройден |
| 7 | Удаление задачи                                | HTTP 204                     | HTTP 204                     | Пройден  |
| 8 | Получение несуществующей задачи                | HTTP 404                     | HTTP 404                     | Пройден  |
| 9 | Получение статистики                           | HTTP 200, объект статистики  | HTTP 200, объект статистики  | Пройден  |
| 10| Обновление с конфликтом версии                 | HTTP 409                     | HTTP 409                     | Пройден  |

Рисунок 3 — Скриншот Swagger UI
[Вставить скриншот интерфейса Swagger по адресу http://127.0.0.1:5050/swagger/]

Рисунок 4 — Скриншот веб-интерфейса TodoApp
[Вставить скриншот веб-интерфейса по адресу http://127.0.0.1:5050/]

Все десять тестовых сценариев пройдены успешно. Приложение корректно обрабатывает валидные и невалидные запросы, возвращает соответствующие HTTP-коды и сообщения об ошибках.

---

5 Оценка результатов

По итогам реализации проведена оценка соответствия разработанного приложения требованиям технического задания.

Реализованные функциональные возможности:
— полный CRUD для пользователей (создание, получение, обновление, удаление) с валидацией данных;
— полный CRUD для задач с привязкой к пользователям, автоматическим управлением временем завершения и оптимистичной блокировкой;
— получение статистики по задачам с фильтрацией по пользователю и временному диапазону;
— пагинация списков через параметры limit и offset;
— интерактивная Swagger-документация с автоматической генерацией из аннотаций кода;
— веб-интерфейс для визуального управления задачами и пользователями.

Реализованные нефункциональные требования:
— graceful shutdown с таймаутом 30 секунд;
— пул соединений с PostgreSQL через библиотеку pgx;
— структурированное логирование через zap с записью в файл и stdout;
— поддержка CORS;
— контейнеризация через Docker Compose (PostgreSQL, миграции, port-forwarder, Swagger).

Все функциональные и нефункциональные требования, определённые в техническом задании, реализованы в полном объёме.

---

ЗАКЛЮЧЕНИЕ

В ходе производственной практики разработано RESTful веб-приложение TodoApp для управления задачами на языке программирования Go с применением многослойной архитектуры.

Выполнены следующие работы:
— реализована серверная часть приложения с разделением на слои Transport, Service и Repository;
— реализованы функциональные модули: users (управление пользователями), tasks (управление задачами), statistics (статистика) и web (веб-интерфейс);
— обеспечено хранение данных в СУБД PostgreSQL с управлением схемой через миграции (golang-migrate);
— реализован REST API с 11 эндпоинтами и автоматической Swagger-документацией;
— применены паттерны проектирования: Dependency Injection, Repository Pattern, Optimistic Locking;
— обеспечена контейнеризация компонентов через Docker Compose;
— проведено функциональное тестирование, все 10 тестовых сценариев пройдены успешно.

В процессе выполнения практики получены навыки:
— разработки серверных приложений на языке Go;
— проектирования REST API и работы со спецификацией OpenAPI;
— работы с СУБД PostgreSQL через драйвер pgx;
— применения многослойной архитектуры и принципов SOLID;
— контейнеризации приложений с помощью Docker и Docker Compose;
— использования системы контроля версий Git.

Исходный код приложения размещён в открытом репозитории GitHub.

---

СПИСОК ИСПОЛЬЗОВАННЫХ ИСТОЧНИКОВ

1. Донован, А. Язык программирования Go / А. Донован, Б. Керниган. — М. : Вильямс, 2021. — 432 с.
2. Мартин, Р. Чистая архитектура. Искусство разработки программного обеспечения / Р. Мартин. — СПб. : Питер, 2018. — 352 с.
3. Fielding, R. T. Architectural Styles and the Design of Network-based Software Architectures : дис. ... д-ра философии / R. T. Fielding. — University of California, Irvine, 2000. — 162 p.
4. PostgreSQL 18 Documentation [Электронный ресурс]. — Режим доступа: https://www.postgresql.org/docs/18/ (дата обращения: 01.06.2026).
5. The Go Programming Language Specification [Электронный ресурс]. — Режим доступа: https://go.dev/ref/spec (дата обращения: 01.06.2026).
6. Docker Documentation [Электронный ресурс]. — Режим доступа: https://docs.docker.com/ (дата обращения: 01.06.2026).
7. OpenAPI Specification v3.0 [Электронный ресурс]. — Режим доступа: https://swagger.io/specification/ (дата обращения: 01.06.2026).
8. ГОСТ 7.32-2017. Отчёт о научно-исследовательской работе. Структура и правила оформления. — М. : Стандартинформ, 2017.
9. ГОСТ Р 7.0.5-2008. Библиографическая ссылка. Общие требования и правила составления. — М. : Стандартинформ, 2008.
10. Gamma, E. Design Patterns: Elements of Reusable Object-Oriented Software / E. Gamma, R. Helm, R. Johnson, J. Vlissides. — Addison-Wesley, 1994. — 395 p.
11. Фаулер, М. Шаблоны корпоративных приложений / М. Фаулер. — М. : Вильямс, 2020. — 544 с.
12. Richardson, L. RESTful Web Services / L. Richardson, S. Ruby. — O'Reilly Media, 2007. — 454 p.
13. pgx — PostgreSQL Driver and Toolkit for Go [Электронный ресурс]. — Режим доступа: https://github.com/jackc/pgx (дата обращения: 01.06.2026).
14. Zap — Blazing fast, structured, leveled logging in Go [Электронный ресурс]. — Режим доступа: https://github.com/uber-go/zap (дата обращения: 01.06.2026).
15. golang-migrate — Database migrations written in Go [Электронный ресурс]. — Режим доступа: https://github.com/golang-migrate/migrate (дата обращения: 01.06.2026).

---

ПРИЛОЖЕНИЕ А
(обязательное)

Исходный код main.go

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

        _ "github.com/milas1221/golang-todoapp/docs"
    )

    func main() {
        cfg := core_config.NewConfigMust()
        time.Local = cfg.TimeZone

        ctx, cancel := signal.NotifyContext(
            context.Background(),
            syscall.SIGINT, syscall.SIGTERM,
        )
        defer cancel()

        logger, err := core_logger.NewLogger(core_logger.NewConfigMust())
        if err != nil {
            fmt.Println("failed to init application logger:", err)
            os.Exit(1)
        }
        defer logger.Close()

        logger.Debug("application time zone", zap.Any("zone", time.Local))

        logger.Debug("initializing postgres connection pool")
        postgresPool, err := core_pgx_pool.NewPool(
            ctx,
            core_pgx_pool.NewConfigMust(),
        )
        if err != nil {
            logger.Fatal("failed to init postgres connection pool", zap.Error(err))
        }
        defer postgresPool.Close()

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

        apiVersionRouterV1 := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
        apiVersionRouterV1.AddRoutes(usersTransportHTTP.Routes()...)
        apiVersionRouterV1.AddRoutes(tasksTransportHTTP.Routes()...)
        apiVersionRouterV1.AddRoutes(statisticsTransportHTTP.Routes()...)

        httpServer.RegisterAPIRouters(apiVersionRouterV1)
        httpServer.RegisterRoutes(webTransportHTTP.Routes()...)
        httpServer.RegisterSwagger()

        if err := httpServer.Run(ctx); err != nil {
            logger.Error("HTTP server run error", zap.Error(err))
        }
    }

---

ПРИЛОЖЕНИЕ Б
(обязательное)

Исходный код доменных моделей

Файл internal/core/domain/task.go:

    package domain

    import (
        "fmt"
        "time"

        "github.com/google/uuid"
        core_errors "github.com/milas1221/golang-todoapp/internal/core/errors"
    )

    type Task struct {
        ID           uuid.UUID
        Version      int
        Title        string
        Description  *string
        Completed    bool
        CreatedAt    time.Time
        CompletedAt  *time.Time
        AuthorUserID uuid.UUID
    }

    func CreateTask(
        title string,
        description *string,
        authorUserID uuid.UUID,
    ) Task {
        var (
            id                     = uuid.New()
            version                = 1
            completed              = false
            createdAt              = time.Now()
            completedAt *time.Time = nil
        )
        return NewTask(id, version, title, description,
            completed, createdAt, completedAt, authorUserID)
    }

    func (t *Task) Validate() error {
        titleLen := len([]rune(t.Title))
        if titleLen < 1 || titleLen > 100 {
            return fmt.Errorf(
                "invalid `Title` len: %d: %w",
                titleLen, core_errors.ErrInvalidArgument,
            )
        }
        if t.Description != nil {
            descriptionLen := len([]rune(*t.Description))
            if descriptionLen < 1 || descriptionLen > 1000 {
                return fmt.Errorf(
                    "invalid `Description` len: %d: %w",
                    descriptionLen, core_errors.ErrInvalidArgument,
                )
            }
        }
        if t.Completed {
            if t.CompletedAt == nil {
                return fmt.Errorf(
                    "`CompletedAt` can't be nil if Completed==true: %w",
                    core_errors.ErrInvalidArgument,
                )
            }
            if t.CompletedAt.Before(t.CreatedAt) {
                return fmt.Errorf(
                    "`CompletedAt` can't be before `CreatedAt`: %w",
                    core_errors.ErrInvalidArgument,
                )
            }
        } else {
            if t.CompletedAt != nil {
                return fmt.Errorf(
                    "`CompletedAt` must be nil if Completed==false: %w",
                    core_errors.ErrInvalidArgument,
                )
            }
        }
        return nil
    }

Файл internal/core/domain/user.go:

    package domain

    import (
        "fmt"
        "regexp"

        "github.com/google/uuid"
        core_errors "github.com/milas1221/golang-todoapp/internal/core/errors"
    )

    type User struct {
        ID          uuid.UUID
        Version     int
        FullName    string
        PhoneNumber *string
    }

    func CreateUser(
        fullName string,
        phoneNumber *string,
    ) User {
        var (
            id      = uuid.New()
            version = 1
        )
        return NewUser(id, version, fullName, phoneNumber)
    }

    func (u *User) Validate() error {
        fullNameLen := len([]rune(u.FullName))
        if fullNameLen < 3 || fullNameLen > 100 {
            return fmt.Errorf(
                "invalid `FullName` len: %d: %w",
                fullNameLen, core_errors.ErrInvalidArgument,
            )
        }
        if u.PhoneNumber != nil {
            phoneNumberLen := len([]rune(*u.PhoneNumber))
            if phoneNumberLen < 10 || phoneNumberLen > 15 {
                return fmt.Errorf(
                    "invalid `PhoneNumber` len: %d: %w",
                    phoneNumberLen, core_errors.ErrInvalidArgument,
                )
            }
            re := regexp.MustCompile(`^\+[0-9]+$`)
            if !re.MatchString(*u.PhoneNumber) {
                return fmt.Errorf(
                    "invalid `PhoneNumber` format: %w",
                    core_errors.ErrInvalidArgument,
                )
            }
        }
        return nil
    }

---

ПРИЛОЖЕНИЕ В
(обязательное)

Файл docker-compose.yaml

    services:
      todoapp:
        build:
          context: ${PROJECT_ROOT}
          dockerfile: ${PROJECT_ROOT}/cmd/todoapp/Dockerfile
        container_name: todoapp
        env_file:
          - ${PROJECT_ROOT}/.env
        environment:
          - HTTP_ADDR=:5050
          - POSTGRES_HOST=todoapp-postgres
          - REDIS_HOST=todoapp-redis
          - LOGGER_FOLDER=/app/out/logs
        volumes:
          - ${PROJECT_ROOT}/out/logs:/app/out/logs
        ports:
          - "5050:5050"
        restart: always

      todoapp-postgres:
        image: postgres:18.1-bookworm
        container_name: todoapp-env-postgres
        environment:
          - POSTGRES_USER=${POSTGRES_USER}
          - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
          - POSTGRES_DB=${POSTGRES_DB}
        volumes:
          - ${PROJECT_ROOT}/out/pgdata:/var/lib/postgresql
        restart: always

      todoapp-postgres-migrate:
        image: migrate/migrate:v4.19.1
        volumes:
          - ${PROJECT_ROOT}/migrations:/migrations

      port-forwarder:
        image: alpine/socat:1.8.0.3
        container_name: todoapp-env-port-forwarder
        command: >-
          tcp-listen:5433,fork,reuseaddr
          tcp-connect:todoapp-postgres:5432
        ports:
          - "127.0.0.1:5433:5433"

      swagger:
        build:
          context: ${PROJECT_ROOT}
          dockerfile_inline: |
            FROM golang:1.24.1-alpine
            RUN apk add --no-cache git
            RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6
            WORKDIR /code
            ENTRYPOINT ["swag"]
        volumes:
          - ${PROJECT_ROOT}:/code
