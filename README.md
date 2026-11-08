# TripGo — лабораторная работа 1

HTTP-сервис для создания, получения и завершения поездок. Данные хранятся в PostgreSQL. Создание и завершение поездки записывают историю статусов в той же транзакции.

## Требования

- Go 1.26.3 или новее, Git, GNU Make.
- Docker и настроенный `tripgoctl` из [course-infra](https://github.com/course-go-autumn-2026/course-infra).
- Для примеров запросов: `curl` и `jq`.
- На Windows команды ниже выполняются в терминале WSL. VS Code следует открыть в режиме WSL.

## Запуск

Все команды выполняются из корня репозитория.

```bash
git clone https://github.com/Zvoook/TripGo_Danilov.git
cd TripGo_Danilov
git checkout homework/01
go mod download
tripgoctl cluster start
tripgoctl environment start
tripgoctl connect
```

`tripgoctl connect` обеспечивает доступ к PostgreSQL. Если команда занимает терминал, оставьте её работать и откройте второй терминал в той же папке.

В `.env` сохраните `DATABASE_URL`, созданный `tripgoctl`: адрес и порт зависят от окружения. Не заменяйте его адресом из `.env.example`.

Добавьте в `.env` недостающие настройки:

```dotenv
HTTP_ADDR=:8080
HTTP_READ_HEADER_TIMEOUT=5s
HTTP_READ_TIMEOUT=10s
HTTP_WRITE_TIMEOUT=15s
HTTP_IDLE_TIMEOUT=60s
LOG_LEVEL=info
SHUTDOWN_TIMEOUT=10s
DATABASE_MAX_CONNS=10
DATABASE_MIN_CONNS=2
DATABASE_MAX_CONN_LIFETIME=30m
DATABASE_CONNECT_TIMEOUT=5s
DATABASE_QUERY_TIMEOUT=3s
```

Затем:

```bash
make migrate
make run
```

`Makefile` читает `.env` и передаёт переменные приложению. При прямом запуске бинарного файла переменные нужно экспортировать самостоятельно. При недоступной БД или неправильной обязательной настройке приложение завершается с ошибкой.

Остановить сервис можно через `Ctrl+C`. Также обрабатывается `SIGTERM`. После остановки сервиса окружение можно остановить командой `tripgoctl environment stop`.

## Переменные окружения

Все перечисленные переменные обязательны. Значения ниже — пример настройки, а не встроенные значения по умолчанию.

| Переменная | Пример | Назначение |
|---|---|---|
| `HTTP_ADDR` | `:8080` | Адрес HTTP-сервера |
| `HTTP_READ_HEADER_TIMEOUT` | `5s` | Время чтения заголовков |
| `HTTP_READ_TIMEOUT` | `10s` | Время чтения запроса |
| `HTTP_WRITE_TIMEOUT` | `15s` | Время записи ответа |
| `HTTP_IDLE_TIMEOUT` | `60s` | Ожидание следующего запроса на соединении |
| `LOG_LEVEL` | `info` | Проверяется конфигом; настройка уровней логирования относится к ЛР2 |
| `SHUTDOWN_TIMEOUT` | `10s` | Время ожидания завершения HTTP-запросов при остановке |
| `DATABASE_URL` | Из `tripgoctl` | Строка подключения к PostgreSQL |
| `DATABASE_MAX_CONNS` | `10` | Максимальное число соединений, больше нуля |
| `DATABASE_MIN_CONNS` | `2` | Минимальное число соединений, от нуля до максимума |
| `DATABASE_MAX_CONN_LIFETIME` | `30m` | Время жизни соединения |
| `DATABASE_CONNECT_TIMEOUT` | `5s` | Таймаут подключения и стартового Ping |
| `DATABASE_QUERY_TIMEOUT` | `3s` | Таймаут запросов, операций с поездкой и отката транзакции |

`LOG_LEVEL` принимает `debug`, `info`, `warn`, `error`. Длительности должны быть положительными. Переменные следующих лабораторных в `.env.example` сейчас не используются. `.env` и `.tripgo/` исключены из Git.

## HTTP API

| Метод | Путь | Результат |
|---|---|---|
| GET | `/health` | `200`, процесс работает; БД не проверяется |
| GET | `/ready` | `200`, БД доступна; иначе `503` |
| POST | `/api/v1/trips` | Создание поездки, `201` и заголовок `Location` |
| GET | `/api/v1/trips/{tripId}` | Получение поездки, `200` |
| POST | `/api/v1/trips/{tripId}/finish` | Завершение поездки, `200` |

Цена задаётся целым числом рублей и может быть нулевой. Идентификаторы пользователя и водителя — ненулевые UUID. Широта от -90 до 90, долгота от -180 до 180.

Ошибки операций с поездками возвращаются как `application/problem+json`:

- `400 invalid_request` — неправильные входные данные;
- `404 trip_not_found` — поездка не найдена;
- `409 driver_busy` — у водителя уже есть активная поездка;
- `409 trip_completed` — поездка уже завершена;
- `500 internal_error` — внутренняя ошибка без раскрытия SQL и деталей БД.

### Пример запросов

```bash
curl -i http://localhost:8080/health
curl -i http://localhost:8080/ready

RESPONSE=$(curl -sS -X POST http://localhost:8080/api/v1/trips \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"5cb72c04-7650-45c9-a79b-bcdba0631e0c","driver_id":"8860b315-ec86-42eb-a17c-7c163d721ff5","start_point":{"latitude":59.9398,"longitude":30.3146},"end_point":{"latitude":59.9290,"longitude":30.3626},"price":1450}')
printf '%s\n' "$RESPONSE" | jq
TRIP_ID=$(printf '%s\n' "$RESPONSE" | jq -r '.id')

curl -i "http://localhost:8080/api/v1/trips/$TRIP_ID"
curl -i -X POST "http://localhost:8080/api/v1/trips/$TRIP_ID/finish"
curl -i -X POST "http://localhost:8080/api/v1/trips/$TRIP_ID/finish"
```

Первое завершение возвращает `200`, повторное — `409 trip_completed`. Если создание вернуло `driver_busy`, завершите прежнюю поездку этого водителя или используйте другой `driver_id`.

## Команды разработки

```bash
go build ./...
go vet ./...
make test
make generate
make build
```

`make test` запускает `go test -race ./...`. Автоматические тесты в ЛР1 не добавлены, поэтому вывод содержит `[no test files]`; это не проверка поведения API.

`make generate` создаёт типы и серверный интерфейс пяти операций ЛР1 в `internal/generated/api.gen.go`. Сгенерированный файл хранится в Git и не редактируется вручную. Версии генератора и goose закреплены в `go.mod` и `go.sum`.

`make build` создаёт `bin/trip-service`. Контракт находится в `contracts/openapi/trip-service.openapi.yaml`.

## Миграции

`make migrate` применяет обе миграции: `trips` и `trip_status_history`.

`make migrate-down` откатывает одну последнюю миграцию. Два вызова откатят обе. Откат удаляет таблицы вместе с данными: проверять его следует на отдельной тестовой БД или схеме.

## Структура

- `cmd/trip-service` — сборка зависимостей, запуск и остановка сервера.
- `internal/config` — чтение и проверка env.
- `internal/httpapi` — обработчики, чтение JSON, ответы и HTTP-ошибки.
- `internal/trip` — модель, бизнес-операции и репозиторий.
- `internal/postgres` — пул, менеджер транзакций и выбор исполнителя запросов.
- `internal/generated` — сгенерированные типы и маршруты.
- `migrations` — SQL-миграции goose.
- `contracts` — контракты курса.

## Решения

Используется уровень изоляции **Read Committed**. Для наших операций его достаточно вместе с уникальным индексом и блокировкой строки при завершении.

Менеджер транзакций `Do` открывает транзакцию и сохраняет её в контексте. Репозиторий получает исполнителя из контекста: внутри `Do` это транзакция, вне `Do` — пул. Вложенный `Do` использует уже открытую транзакцию.

При успехе выполняется `Commit`. При ошибке или панике отложенный `Rollback` отменяет изменения. Ошибка возвращается вызывающему коду, паника продолжает распространяться. Откат получает отдельный таймаут и не зависит от отмены исходного контекста.

Создание поездки и запись `NULL → active` в историю выполняются в одной транзакции. Ошибка второй вставки отменяет первую.

Частичный уникальный индекс `one_active_trip_per_driver_idx` запрещает несколько активных поездок одному водителю. Ошибка PostgreSQL `23505` именно для этого индекса превращается в `driver_busy`.

При завершении `SELECT FOR UPDATE` блокирует строку до конца транзакции. После проверки статуса сервис меняет поездку и записывает `active → completed` в историю. Параллельный запрос после ожидания увидит `completed` и получит `409`; время завершения не изменится.
