# Polling Service

Высокопроизводительный REST API для создания опросов и голосования с защитой от повторных голосов.

## Возможности

- Создание опросов с несколькими вариантами ответов
- Голосование за вариант с атомарным обновлением счетчика
-  Защита от повторного голосования (уникальный индекс в БД)
- Получение опросов по ID и списком с пагинацией
- Конкурентная безопасность через транзакции PostgreSQL
- Graceful shutdown
- Структурированные логи (slog)
- UUID вместо числовых ID
- Чистая архитектура (handler → service → repository)

## Технологии

- **Go** — основной язык
- **PostgreSQL** — хранение данных
- **pgx** — драйвер для PostgreSQL
- **chi** — маршрутизация HTTP
- **golang-migrate** — миграции БД
- **Docker Compose** — быстрый запуск окружения

## Архитектура

```
polling-service/
├── cmd/
│ ├── api/ # Точка входа HTTP сервера
│ └── migrate/ # Утилита для миграций БД
├── internal/
│ ├── config/ # Конфигурация
│ ├── domain/ # Сущности и бизнес-правила
│ ├── handler/ # HTTP обработчики
│ ├── repository/ # Работа с БД
│ └── service/ # Бизнес-логика
└── migrations/ # SQL миграции
```

## Быстрый старт

### 1. Клонируйте репозиторий

```bash
git clone https://github.com/ВАШ_ЛОГИН/polling-service.git
cd polling-service
```

### 2. Настройка
1. Скопируйте файл с примером переменных:
```bash
cp .env.example .env
```
2. Отредактируйте `.env` под свои нужды:
```
DB_USER=my_user
DB_PASSWORD=my_secure_password
DB_NAME=my_database
```

### 3. Запустите окружение и приложение

```bash
# Полный старт (Docker + миграции + приложение)
make start

# Или по шагам:
make docker-up      # Поднять PostgreSQL
make migrate-up     # Применить миграции
make run            # Запустить сервер
```

### 4. Остановка

```bash
make docker-down    # Остановить контейнеры
```

### Другие полезные команды

```bash
make help           # Показать все доступные команды
make reset          # Полностью пересоздать БД
make test           # Запустить тесты
make build          # Собрать бинарник
```

## API Эндпоинты

| Метод         | Путь              | Описание      |
| ------------- | -------------     | ------------- |
| `POST`          | `/polls`            | Создать опрос  |
| `GET`           | `/polls`            | Список опросов (с пагинацией)  |
| `GET`           | `/polls/{id}`       | Получить опрос по ID  |
| `POST`          | `/polls/{id}/vote`  | 	Проголосовать  |
| `GET`           | `/polls/votes`      | Получить голоса  |

### Примеры запросов

#### Создание опроса:
```bash
curl -X POST http://localhost:8080/api/polls \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Лучший язык программирования?",
    "options": ["Go", "Python", "Rust", "JavaScript"]
  }'
```

#### Получение опроса по ID:
```bash
curl http://localhost:8080/api/polls/{poll_id}
```

#### Голосование:

```bash
curl -X POST http://localhost:8080/api/polls/{poll_id}/votes \
  -H "Content-Type: application/json" \
  -d '{
    "option_id": "{option_id}",
    "user_id": "user123"
  }'
```

#### Получение списка опросов:

```bash
curl http://localhost:8080/api/polls?page=1&page_size=10
```

### Пример ответа
#### Создание опроса (201 Created):

```json
{
  "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "title": "Лучший язык программирования?",
  "options": [
    {
      "id": "a1b2c3d4-...",
      "poll_id": "f47ac10b-...",
      "text": "Go",
      "votes_count": 0
    },
    {
      "id": "e5f6g7h8-...",
      "poll_id": "f47ac10b-...",
      "text": "Python",
      "votes_count": 0
    }
  ],
  "created_at": "2026-07-01T15:30:00Z"
}
```
#### Голосование (200 OK):

```json
{
  "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "title": "Лучший язык программирования?",
  "options": [
    {
      "id": "a1b2c3d4-...",
      "poll_id": "f47ac10b-...",
      "text": "Go",
      "votes_count": 1
    },
    {
      "id": "e5f6g7h8-...",
      "poll_id": "f47ac10b-...",
      "text": "Python",
      "votes_count": 0
    }
  ],
  "created_at": "2026-07-01T15:30:00Z"
}
```
## Защита от повторного голосования
Система использует уникальный индекс `(poll_id, user_id)` в **PostgreSQL**, что гарантирует:

- Один пользователь может проголосовать только один раз в каждом опросе

- Конкурентная безопасность через атомарные операции

- Возврат статуса `409 Conflict` при попытке повторного голосования

### Коды ответов
|Код|	Описание|
|----|----------|
|200|	Успешный запрос|
|201|	Опрос создан|
|400|	Некорректный запрос (валидация)|
|404|	Опрос или вариант не найден|
|409|	Пользователь уже голосовал|
|500|	Внутренняя ошибка сервера|

## Дорожная карта
- Итерация 1: Базовый CRUD для опросов
- Итерация 2: Голосование и защита от накруток
- Итерация 3: Кэширование в Redis
- Итерация 4: WebSocket для результатов в реальном времени

## Зависимости
```bash
go get -u github.com/go-chi/chi/v5
go get -u github.com/go-chi/cors
go get -u github.com/jackc/pgx/v5
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/joho/godotenv
go get -u github.com/google/uuid
```
## Лицензия
MIT

## Автор
Popodada3221 - [GitHub](https://github.com/Popodada3221)