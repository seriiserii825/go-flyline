# 01. `GET /api/v1/health`

Сейчас роут лежит в `cmd/flyline/main.go`, и из-за этого `main` импортирует gin. По правилу из TODO gin допустим
только в `handler.go` / `middleware.go` / `routes.go` / `cookie.go` и в `platform/httpserver`, `platform/middleware`.
Значит, health переезжает в `platform/httpserver`.

Проверку БД сейчас не делаем (пула ещё нет), но оставляем под неё место: в разделе 2 в `main` добавится одна строка.

## Шаги

- [ ] Создать `internal/platform/httpserver/health.go`
- [ ] В нём тип проверки: `type HealthCheck func(ctx context.Context) error` (`nil` — всё хорошо)
- [ ] В `Server` (`server.go`) добавить поле `healthChecks map[string]HealthCheck`
- [ ] В `New`: инициализировать map (`map[string]HealthCheck{}`), иначе запись в nil-map вызовет панику
- [ ] Метод `func (s *Server) AddHealthCheck(name string, check HealthCheck)` записывает проверку в map
- [ ] Метод-хендлер `func (s *Server) handleHealth(c *gin.Context)`:
  - [ ] `context.WithTimeout(c.Request.Context(), 2*time.Second)` + `defer cancel()`: зависшая БД не должна
        вешать health
  - [ ] прогнать все проверки, ошибки собрать в map `name → err.Error()`, залогировать через `s.log.Warn`
  - [ ] если есть упавшие → `503` и `{"status":"unavailable","checks":{...}}`
  - [ ] иначе → `200` и `{"status":"ok"}`
- [ ] В `New`: заменить `return &Server{...}` на `s := &Server{...}`, затем `s.api.GET("/health", s.handleHealth)`,
      затем `return s`
- [ ] Из `main.go` убрать роут и импорты `gin`, `net/http`
- [ ] `go build ./...` и `go vet ./...`

## Проверка

- [ ] `make run`
- [ ] `curl -i localhost:8080/api/v1/health` → `200 {"status":"ok"}`
- [ ] (по желанию) временно добавить в `main` проверку, которая всегда падает:
      `srv.AddHealthCheck("fake", func(context.Context) error { return errors.New("boom") })` → должен прийти
      `503`, потом удалить

## Потом (раздел 2)

- [ ] В `main`: `srv.AddHealthCheck("db", pool.Ping)`. Сигнатура `pool.Ping(ctx) error` подходит под `HealthCheck`
      как есть
