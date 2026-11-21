# TripGo — лабораторная работа 3

HTTP-сервис для управления поездками и хранения маршрута в PostgreSQL. Если активная поездка давно не присылала координаты, сервис запрашивает их через Push Service.

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
git checkout homework/03
go mod download
tripgoctl cluster start
DOCKER_BUILDKIT=1 tripgoctl environment start
tripgoctl connect
```

`tripgoctl connect` обеспечивает доступ к PostgreSQL. Если команда занимает терминал, оставьте её работать и откройте второй терминал в той же папке.

В `.env` сохраните `DATABASE_URL` и `PUSH_HTTP_URL`, созданные `tripgoctl`: адреса и порты зависят от окружения.

Остальные настройки необязательны. При необходимости переопределите в `.env` значения по умолчанию:

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
OPERATION_TIMEOUT=10s
```

Затем:

```bash
make migrate
make run
```

`Makefile` читает `.env` и передаёт переменные приложению. При прямом запуске бинарного файла переменные нужно экспортировать самостоятельно. При недоступной БД или неправильной обязательной настройке приложение завершается с ошибкой.

Остановить сервис можно через `Ctrl+C`. Также обрабатывается `SIGTERM`. После остановки сервиса окружение можно остановить командой `tripgoctl environment stop`.

## Переменные окружения

Обязательны `DATABASE_URL` и `PUSH_HTTP_URL`. Для остальных переменных ниже указаны встроенные значения по умолчанию. Пустое значение необязательной переменной также использует значение по умолчанию; неправильное непустое значение приводит к ошибке запуска.

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `HTTP_ADDR` | `:8080` | Адрес HTTP-сервера |
| `HTTP_READ_HEADER_TIMEOUT` | `5s` | Время чтения заголовков |
| `HTTP_READ_TIMEOUT` | `10s` | Время чтения запроса |
| `HTTP_WRITE_TIMEOUT` | `15s` | Время записи ответа |
| `HTTP_IDLE_TIMEOUT` | `60s` | Ожидание следующего запроса на соединении |
| `LOG_LEVEL` | `info` | Минимальный уровень JSON-логов |
| `SHUTDOWN_TIMEOUT` | `10s` | Общий бюджет остановки сервиса |
| `DATABASE_URL` | Из `tripgoctl` | Строка подключения к PostgreSQL |
| `DATABASE_MAX_CONNS` | `10` | Максимальное число соединений, больше нуля |
| `DATABASE_MIN_CONNS` | `2` | Минимальное число соединений, от нуля до максимума |
| `DATABASE_MAX_CONN_LIFETIME` | `30m` | Время жизни соединения |
| `DATABASE_CONNECT_TIMEOUT` | `5s` | Таймаут подключения и стартового Ping |
| `DATABASE_QUERY_TIMEOUT` | `3s` | Таймаут одного SQL-запроса и отдельный бюджет отката |
| `OPERATION_TIMEOUT` | `10s` | Общий бюджет создания или завершения поездки, включая транзакцию |

`LOG_LEVEL` принимает `debug`, `info`, `warn`, `error`. Длительности должны быть положительными. `.env` и `.tripgo/` исключены из Git. Настройки опроса перечислены ниже.

## HTTP API

| Метод | Путь | Результат |
|---|---|---|
| GET | `/health` | `200`, процесс работает; БД не проверяется |
| GET | `/ready` | `200`, БД доступна; иначе `503` |
| POST | `/api/v1/trips` | Создание поездки, `201` и заголовок `Location` |
| GET | `/api/v1/trips/{tripId}` | Получение поездки, `200` |
| POST | `/api/v1/trips/{tripId}/finish` | Завершение поездки, `200` |
| POST | `/api/v1/trips/{tripId}/positions` | Сохранение координаты активной поездки, `201` |
| GET | `/api/v1/trips/{tripId}/positions` | Маршрут в порядке возрастания `recorded_at`, `200` |

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
curl -i -X POST "http://localhost:8080/api/v1/trips/$TRIP_ID/positions" \
  -H 'Content-Type: application/json' \
  -d "{\"coordinates\":{\"latitude\":59.93,\"longitude\":30.33},\"recorded_at\":\"$(date -u +%FT%TZ)\"}"
curl -i "http://localhost:8080/api/v1/trips/$TRIP_ID/positions"
curl -i -X POST "http://localhost:8080/api/v1/trips/$TRIP_ID/finish"
curl -i -X POST "http://localhost:8080/api/v1/trips/$TRIP_ID/finish"
```

Первое завершение возвращает `200`, повторное — `409 trip_completed`. Если создание вернуло `driver_busy`, завершите прежнюю поездку этого водителя или используйте другой `driver_id`.

## Команды разработки

```bash
go build ./...
go vet ./...
make lint
make test
make generate
make build
```

`make test` запускает unit-тесты с детектором гонок и печатает покрытие. PostgreSQL для них не требуется.

Для интеграционных тестов поднимите окружение ЛР3 и проверьте адрес базы через `tripgoctl connect`. Затем примените миграции и запустите тесты:

```bash
make migrate
make test-integration
```

`make test-integration` читает `DATABASE_URL` из окружения или `.env` и запускает тесты с тегом `integration`. Они проверяют транзакции поездок и координат, выборку устаревших поездок и то, что опоздавшая координата не сдвигает `last_position_at` назад. Каждый сценарий создаёт свои данные и удаляет их после проверки.

`make generate` создаёт типы и серверный интерфейс семи операций в `internal/generated/api.gen.go`. Сгенерированный файл хранится в Git и не редактируется вручную. Версии генератора и goose закреплены в `go.mod` и `go.sum`.

`make build` создаёт `bin/trip-service`. Контракт находится в `contracts/openapi/trip-service.openapi.yaml`.

По замечанию ревью в схемах `Trip` и `TripPosition` свойства перечислены в одном объекте вместо `allOf` с несколькими `additionalProperties: false`. Имена, типы и обязательность полей сохранены: исправлена валидация схемы, которая раньше отвергала корректные ответы.

## Миграции

`make migrate` применяет четыре миграции: таблицы `trips`, `trip_status_history`, `trip_positions` и частичный индекс для поиска устаревших активных поездок. Миграция координат также добавляет `trips.last_position_at`.

`make migrate-down` откатывает одну последнюю миграцию. Откат миграций таблиц удаляет данные: проверять его следует на отдельной тестовой БД или схеме.

## Архитектура

```text
main → container
HTTP → usecase → domain
         ↓ интерфейсы потребителя
репозиторий / менеджер транзакций → PostgreSQL
```

`internal/container` собирает конкретные реализации. Интерфейсы находятся в `contract.go` пакета-потребителя: хендлер объявляет нужный ему метод usecase, а usecase — методы репозитория и транзакций. Домен не зависит от HTTP, базы и OpenTelemetry. Маппинг доменных ошибок находится в `internal/http/domainerror`. Состав маршрутов и поля ответов сохранены. Схемы `Trip` и `TripPosition` перечисляют поля в одном объекте, чтобы `additionalProperties: false` корректно проверял объединённый набор полей.

Модель поездки хранится в `internal/domain/trip` и передаётся между слоями. Отдельный `internal/model` не нужен, пока модель домена совпадает с данными, которые использует usecase. Менеджер транзакций лежит в пакете `internal/infra/postgres` рядом с общим выбором SQL-исполнителя. Обёртки в контейнере записывают бизнес-события после успешного завершения usecase.

Менеджер транзакций передаёт транзакцию в контексте. Создание и завершение поездки записывают историю в той же транзакции. При ошибке выполняется откат. Уникальный индекс защищает водителя от двух активных поездок, а `SELECT FOR UPDATE` защищает завершение от параллельной записи.

## Опрос устаревших поездок

```text
сканер → очередь → воркеры → Push Service
```

Сканер на каждом тике выбирает активные поездки, для которых последняя координата (или начало поездки, если координат ещё не было) старше порога. Чем дольше поездка молчит, тем раньше её задача попадёт в очередь. Повторная задача пропускается, пока предыдущая ждёт или обрабатывается. Успешное сохранение координаты отменяет ожидающую задачу; воркер пропустит её, даже если она всё ещё лежит в канале.

Когда очередь заполнена, новая задача отбрасывается без ожидания: растут счётчики с `result=dropped`, а в логе появляется предупреждение. Если координат так и нет, очередной тик снова проверит поездку. Вызов Push Service выполняется один раз на задачу, с таймаутом. При остановке сервис закрывает входящие HTTP-запросы, останавливает сканер и в пределах общего бюджета даёт воркерам разобрать очередь. Затем закрывает HTTP-клиент и пул БД.

| Переменная | По умолчанию | Назначение |
|---|---|---|
| `POSITION_CHECK_INTERVAL` | `30s` | Период сканирования; задаёт паузу между повторными проверками |
| `POSITION_STALE_AFTER` | `60s` | Возраст координаты, после которого поездка считается устаревшей |
| `POSITION_WORKERS` | `4` | Число горутин, отправляющих команды |
| `POSITION_QUEUE_SIZE` | `100` | Вместимость канала ожидания |
| `PUSH_HTTP_URL` | Из `tripgoctl` | Базовый URL Push Service |
| `PUSH_REQUEST_TIMEOUT` | `1s` | Таймаут одной отправки |
| `SHUTDOWN_TIMEOUT` | `10s` | Общий бюджет остановки сервиса |

## Телеметрия

Выполните `tripgoctl connect` из ветки `homework/03`. Если локальный `.env` сохранился с предыдущей работы, обновите `DATABASE_URL`, `PUSH_HTTP_URL` и `OTEL_EXPORTER_OTLP_ENDPOINT` адресами из этой команды. Для OTel HTTP задайте `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf`.

После `make migrate` и `make run` выполните HTTP-запросы из примера выше. В JSON-логе запроса есть `trace_id`. Откройте Jaeger по адресу из `tripgoctl connect` и найдите этот трейс. Внутри HTTP-спана находятся операции PostgreSQL.

В Grafana выберите Prometheus и найдите метрики по префиксам `http_server_request_duration`, `db_client_operation_duration` и `tripgo_trips_`. Бизнес-счётчики растут после успешного коммита. HTTP-метрики используют шаблон маршрута.

Для пула найдите префиксы `tripgo_stale_checks`, `tripgo_position_requests` и `tripgo_positions_saved`. В Jaeger спан `positions.process` связан со сканированием через span link. Внутри `positions.process` виден отдельный вызов Push Service. Полученные команды `REQUEST_POSITION` видны в `tripgoctl environment logs push-service`.

[График метрик пула под нагрузкой](docs/worker-pool-load.png) построен по экспорту OpenTelemetry. В проверке использовались 200 устаревших поездок, очередь на 10 задач, 4 воркера и задержка Push Service 5 секунд при таймауте запроса 1 секунда. Размер очереди и число занятых воркеров ограничены настройками, переполнение учитывается в `dropped`.

Для вывода трассировки и метрик в терминал используйте:

```bash
make run OTEL_TRACES_EXPORTER=console OTEL_METRICS_EXPORTER=console OTEL_METRIC_EXPORT_INTERVAL=1000
```

`make lint` использует golangci-lint v2 и конфигурацию курса с проверкой зависимостей между слоями.

Задание: [ЛР3 — координаты, очередь и worker pool](https://github.com/course-go-autumn-2026/course/blob/main/homework/tasks/03-positions-and-worker-pool/README.md).
