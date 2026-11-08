# Окружение лабораторной 1

Работайте в VS Code, окно **WSL: Ubuntu**. Папка Linux:
`/mnt/c/Dima/TripGo/TripGo_Danilov` — тот же репозиторий Windows.
Открыть из PowerShell:

```powershell
wsl -d Ubuntu -- bash -lc 'cd /mnt/c/Dima/TripGo/TripGo_Danilov && code .'
```

Команды в терминале Ubuntu / VS Code WSL:

```bash
tripgoctl cluster start
tripgoctl environment start
tripgoctl connect
make env-check
make generate
```

Используется Docker Engine внутри Ubuntu, автозапуск через systemd включён.
Docker Desktop для этого окружения не требуется. Контейнеры видны через
`docker ps` в Ubuntu и расширение Containers в окне WSL.

PostgreSQL поднимает официальный tripgoctl по environment.toml задания 1.
Адрес находится в .env. Не заменяйте его копией .env.example: адрес в примере
условный. В локальный .env добавлены HTTP_ADDR, LOG_LEVEL, SHUTDOWN_TIMEOUT и
параметры пула DATABASE_*. После пересоздания окружения проверьте их наличие.
.env и .tripgo/ исключены из Git.

В WSL установлены Go, GCC, make, Git, curl, jq, psql, gopls и Delve.
Расширения Go, YAML и Containers установлены в Linux VS Code Server.
Форматирование, задачи и конфигурация отладки находятся в .vscode/.
Инструменты oapi-codegen и goose закреплены в go.mod / go.sum.
make generate генерирует только пять операций лабораторной 1.
Контракты в contracts/ скопированы из курса, не редактируются вручную.

Проверено 2026-09-25:

- Go 1.26.3, gopls 0.23.0, Delve 1.27.2.
- tripgoctl main-f40e16445de5641cb4c13e1b07507f49a255ead7.
- kind 0.27.0, Kubernetes 1.32.2, PostgreSQL ready.
- tripgoctl doctor, go mod verify, go build, go vet, make test.
- Повторная генерация API идентична исходному результату; gopls без ошибок.
- PostgreSQL: подключение, запись, goose up/down во временной схеме.
- Реальный запуск временной программы с -race и breakpoint в Delve.

make env-check повторяет проверку инструментов, сборки, БД и миграций.
Временная схема удаляется; таблицы приложения не затрагиваются.

Код HTTP-сервиса и миграции лабораторной ещё предстоит написать.
Сейчас make test сообщает [no test files] для сгенерированного пакета.
make run, make build и отладка сервиса станут доступны после создания
cmd/trip-service; make migrate — после написания миграций.

Источники:

- https://github.com/course-go-autumn-2026/template
- https://github.com/course-go-autumn-2026/course/blob/main/homework/tasks/01-http-and-postgres/README.md
- https://github.com/course-go-autumn-2026/course/blob/main/homework/docs/getting-started.md
- https://github.com/course-go-autumn-2026/course-infra/blob/main/docs/README.md
