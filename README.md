# Golang Todo App

REST API приложение на Go — учебный проект курса. Реализует управление пользователями, задачами и статистикой.

> В проекте 2 git-ветки:
+ `vesrion/base` -> базовая версия (близкая к версии из видео-урока)
+ `version/extended` -> расширенная версия (с Redis-кешом, Caddy web-сервером и гексоганальной архитектурой)

```bash
# Управление проектом
make help
```

## Технологический стек

| Компонент         | Технология                          |
|-------------------|-------------------------------------|
| Язык              | Go 1.22+                            |
| HTTP-фреймворк    | Стандартный `net/http` (без внешних фреймворков) |
| База данных       | PostgreSQL                          |
| Драйвер БД        | `jackc/pgx/v5`                      |
| Логгер            | `go.uber.org/zap`                   |
| Конфигурация      | `kelseyhightower/envconfig`         |
| Валидация         | `go-playground/validator/v10`       |
| Документация API  | Swagger (`swaggo/swag`)             |
| Миграции БД       | golang-migrate                      |
| Деплой            | Docker                              |

---

## Архитектура

Проект следует принципам **чистой архитектуры** (Clean Architecture).
Каждая фича (`users`, `tasks`, `statistics`, `web`) разделена на три слоя:

```
Transport (HTTP Handler)
      │   Декодирует запрос, вызывает сервис, формирует ответ
      ↓
Service (Business Logic)
      │   Валидация, оркестрация вызовов, доменная логика
      ↓
Repository (Data Access)
      └─  SQL-запросы к PostgreSQL, маппинг моделей

Domain (Core)
          Сущности, инварианты, бизнес-правила — без зависимостей
```

Ключевое отличие от обычной слоистой архитектуры — **инверсия зависимостей (DIP)**:
интерфейсы определяются не в реализующем слое, а в потребляющем.

- `TasksRepository` интерфейс живёт в пакете `tasks_service` — сервис владеет контрактом
- `TasksService` интерфейс живёт в пакете `tasks_transport_http` — транспорт владеет контрактом
- `domain` не импортирует ни один другой внутренний пакет — он полностью независим

Благодаря этому зависимости всегда направлены **внутрь**, к домену, а не наружу к инфраструктуре.

**Dependency Injection** (ручная инъекция зависимостей) реализован в `cmd/todoapp/main.go`:
```
Repository → Service → HTTP Handler
```

---

## Структура проекта

```
.
├── cmd/
│   └── todoapp/
│       └── main.go              # Точка входа: инициализация и запуск
├── internal/
│   ├── core/                    # Общие компоненты, не зависящие от фич
│   │   ├── config/              # Общая конфигурация приложения
│   │   ├── domain/              # Доменные сущности: Task, User, Statistics, File, Nullable
│   │   ├── errors/              # Sentinel-ошибки: ErrNotFound, ErrConflict, ErrInvalidArgument
│   │   ├── logger/              # Структурированный логгер (zap) + паттерн «logger in context»
│   │   └── repository/
│   │       └── postgres/pool/   # Интерфейс пула + реализация на pgx
│   │   └── transport/http/
│   │       ├── middleware/      # CORS, RequestID, Logger, Trace, Panic
│   │       ├── request/         # Хелперы: декодирование тела, path/query параметры
│   │       ├── response/        # HTTPResponseHandler, ResponseWriter, ErrorResponse
│   │       ├── server/          # HTTPServer, APIVersionRouter, Route
│   │       └── types/           # Nullable[T] с UnmarshalJSON для PATCH-запросов
│   └── features/                # Бизнес-фичи приложения
│       ├── tasks/               # CRUD задач
│       ├── users/               # CRUD пользователей
│       ├── statistics/          # Статистика по задачам
│       └── web/                 # Отдача статических HTML-страниц
├── migrations/                  # SQL-миграции (golang-migrate)
├── public/                      # Статические файлы (index.html)
├── docs/                        # Автогенерированная Swagger-документация
├── docker-compose.yaml          # Локальная инфраструктура (PostgreSQL, etc.)
├── postman_collection.json      # Коллекция запросов для Postman
└── Makefile                     # Удобные команды для разработки
```

---
## Локальный запуск

#### Предварительные требования
- `Docker и Docker Compose`
- `Go 1.22+`
- `make`

#### Шаги

```bash
# 1. Создать по примеру .env файл
cp .env.example .env

# 2. Выставить недостающие переменные окружения
code .env

# 3. Поднять окружение (PostgreSQL)
make env-up

# 4. Применить миграции БД
make migrate-up

# 5. Открыть порты сервисов окружения
make env-port-forward

# 6. Запустить локально приложение
make todoapp-run
```

После запуска:
- Главная страница: `http://127.0.0.1:5050/`
- Swagger UI: `http://127.0.0.1:5050/swagger/`
- API доступен на `http://127.0.0.1:5050/api/v1/`

## Деплой

```bash
make env-up
make migrate-up
make todoapp-deploy
```

---

## Переменные окружения

| Переменная               | Описание                                      | Пример                                  |
|--------------------------|-----------------------------------------------|-----------------------------------------|
| `TIME_ZONE`              | Часовой пояс (IANA)                           | `Europe/Moscow`                         |
| `LOGGER_LEVEL`           | Уровень логирования                           | `DEBUG`                                 |
| `LOGGER_FOLDER`          | Директория для лог-файлов                     | `out/logs`                              |
| `POSTGRES_HOST`          | Хост PostgreSQL                               | `localhost`                             |
| `POSTGRES_PORT`          | Порт PostgreSQL                               | `5432`                                  |
| `POSTGRES_USER`          | Пользователь БД                               | `todoapp-test-user`                     |
| `POSTGRES_PASSWORD`      | Пароль БД                                     | `todoapp-test-password`                 |
| `POSTGRES_DB`            | Имя базы данных                               | `todoapp`                               |
| `POSTGRES_TIMEOUT`       | Таймаут запроса к БД                          | `5s`                                    |
| `HTTP_ADDR`              | Адрес и порт HTTP-сервера                     | `:5050`                                 |
| `HTTP_SHUTDOWN_TIMEOUT`  | Таймаут graceful shutdown                     | `30s`                                   |
| `HTTP_ALLOWED_ORIGINS`   | Разрешённые CORS origins (через запятую)      | `http://localhost:3000,null`            |
| `PROJECT_ROOT`           | Корень проекта для полных путей               | `/Users/MyUser/projects/golang-todoapp` |

---

## API

### Пользователи `/api/v1/users`

| Метод    | Путь               | Описание                         |
|----------|--------------------|----------------------------------|
| `POST`   | `/users`           | Создать пользователя             |
| `GET`    | `/users`           | Список пользователей (пагинация) |
| `GET`    | `/users/{id}`      | Получить пользователя по ID      |
| `PATCH`  | `/users/{id}`      | Частично обновить пользователя   |
| `DELETE` | `/users/{id}`      | Удалить пользователя             |

### Задачи `/api/v1/tasks`

| Метод    | Путь               | Описание                                        |
|----------|--------------------|--------------------------------------------------|
| `POST`   | `/tasks`           | Создать задачу                                   |
| `GET`    | `/tasks`           | Список задач (пагинация + фильтр по `user_id`)  |
| `GET`    | `/tasks/{id}`      | Получить задачу по ID                            |
| `PATCH`  | `/tasks/{id}`      | Частично обновить задачу                         |
| `DELETE` | `/tasks/{id}`      | Удалить задачу                                   |

### Статистика `/api/v1/statistics`

| Метод  | Путь            | Описание                                                                 |
|--------|-----------------|--------------------------------------------------------------------------|
| `GET`  | `/statistics`   | Статистика задач (фильтры: `user_id`, `from`, `to` в формате YYYY-MM-DD) |

Полная интерактивная документация доступна в **Swagger UI** по адресу `/swagger/`.
