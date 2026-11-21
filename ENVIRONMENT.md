# Окружение лабораторной 2

На Windows выполняйте команды в WSL: Ubuntu из каталога своего клона репозитория. Для VS Code выполните `code .` из этого каталога в WSL.

Нужны Go 1.26.3 или новее, GCC (для `-race`), GNU Make, Git, Docker и `tripgoctl`. Для `make env-check` также нужны `psql`, `gopls` и `dlv`. Docker должен быть доступен из WSL: подойдёт Docker Desktop с интеграцией WSL или Docker Engine внутри Ubuntu.

```bash
go mod download
tripgoctl cluster start
tripgoctl environment start
tripgoctl connect
```

Если `connect` занимает терминал, оставьте его работать и откройте второй терминал в корне репозитория:

```bash
make migrate
make run
```

Сохраните `DATABASE_URL` и `OTEL_EXPORTER_OTLP_ENDPOINT` из вывода `tripgoctl connect` в локальном `.env`. Для адреса OTel HTTP задайте `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf`. В `.env.example` указаны примерные адреса. Используйте значения из `tripgoctl connect`. Остальные настройки имеют значения по умолчанию и переопределяются через `.env`; полный список — в README. Makefile читает `.env` автоматически. При прямом запуске бинарного файла переменные экспортируются вручную.

Сервис реализует создание, получение и завершение поездки, `/health` и `/ready`. Миграции создают таблицы поездок и истории статусов.

```bash
make generate
go build ./...
go vet ./...
make lint
make test
make test-integration
make env-check
```

`make env-check` проверяет инструменты, сборку, БД и миграции во временной схеме. `make generate` генерирует пять операций ЛР1. Версии генератора OpenAPI и goose закреплены в `go.mod` и `go.sum`.

`.env`, `.tripgo/`, `.vscode/` и результаты сборки не хранятся в Git. Контракты находятся в `contracts/`.
