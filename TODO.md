# go-flyline — план

Цель: бэкенд для продажи авиабилетов на Go. Стек: Gin + PostgreSQL + Swagger/OpenAPI (swaggo), на фронте типы
генерируются из OpenAPI через `npx`.

**Этап 1: только авторизация** (регистрация, логин, access + refresh токены, logout).

---

## 0. Инициализация проекта

- [x] `go mod init github.com/seriiserii825/go-flyline`
- [x] Структура папок (по фичам):
  ```
  cmd/flyline/main.go          # только сборка зависимостей и запуск
  internal/
    platform/                  # общая инфраструктура, без бизнес-логики
      config/config.go         # чтение .env (envconfig)
      database/postgres.go     # pgxpool + Ping
      logger/logger.go         # slog
      httpserver/server.go     # gin.New, middleware, graceful shutdown
      httpserver/errors.go     # единый формат ошибок
      middleware/              # cors.go, request_logger.go
    features/
      user/
        model.go               # User (структура из БД)
        dto.go                 # UserResponse (без password_hash)
        errors.go              # ErrNotFound, ErrEmailTaken
        repository.go          # Create, GetByEmail, GetByID
        service.go
        handler.go             # GET /me
        routes.go              # RegisterRoutes(rg, authMW gin.HandlerFunc)
      auth/
        dto.go                 # RegisterRequest, LoginRequest, TokenResponse
        errors.go              # ErrInvalidCredentials, ErrTokenRevoked
        tokens.go              # выпуск/проверка JWT, генерация+хэш refresh
        cookie.go              # set/clear refresh cookie
        repository.go          # refresh_tokens
        service.go             # Register, Login, Refresh, Logout, LogoutAll
        handler.go
        middleware.go          # AuthRequired
        routes.go              # RegisterRoutes(rg)
  migrations/
  docs/                        # генерируется swag, руками не править
  ```

  Правила:
  - `gin` импортируется только в `handler.go`, `middleware.go`, `routes.go`, `cookie.go` и в `platform/httpserver`
  - импорты только `auth → user`, обратно нельзя

- [x] Добавить `.env.example` (без секретов) — `.gitignore` его уже пропускает
- [x] Переменные в `.env`: `HTTP_PORT`, `ADMINER_PORT`, `GIN_MODE`, `DATABASE_URL`, `JWT_SECRET`, `JWT_ACCESS_TTL=15m`,
      `JWT_REFRESH_TTL=720h`
- [x] Конфиг через `envconfig` (`internal/platform/config`)
- [x] Валидация конфига при старте (`GIN_MODE`, длина `JWT_SECRET` ≥ 32, `JWT_REFRESH_TTL` > `JWT_ACCESS_TTL`)

## 1. Gin: установка и настройка

- [x] `go get github.com/gin-gonic/gin`
- [ ] `internal/platform/httpserver`: создать `gin.New()` + `gin.Recovery()` + свой логгер (slog)
- [x] Режим через env: `GIN_MODE=debug|release` (`gin.SetMode(cfg.GinMode)`)
- [ ] Группа роутов `/api/v1`
- [ ] `GET /api/v1/health` → `{"status":"ok"}` (проверка, что сервер и БД живы)
- [ ] CORS для фронта: `go get github.com/gin-contrib/cors` (разрешить origin фронта, `AllowCredentials: true` — нужно
      для cookie)
- [ ] Единый формат ошибок: `{"error": {"code": "...", "message": "..."}}`
- [ ] Graceful shutdown (`http.Server` + `signal.NotifyContext`)
- [ ] Цель `run` в Makefile запускает сервер, проверить `curl localhost:$HTTP_PORT/api/v1/health`

## 2. База данных

- [ ] `go get github.com/jackc/pgx/v5` (пул `pgxpool`)
- [ ] `internal/platform/database`: подключение + `Ping` при старте
- [ ] `make migrate-create seq=create_users`:
  - `id uuid pk`, `email text unique not null`, `password_hash text not null`, `created_at`, `updated_at`
- [ ] `make migrate-create seq=create_refresh_tokens`:
  - `id uuid pk`, `user_id uuid fk → users on delete cascade`, `token_hash text unique not null`,
    `expires_at timestamptz`, `revoked_at timestamptz null`, `created_at`, `user_agent text`, `ip text`
  - индекс на `user_id`
- [ ] `make migrate-up`, проверить таблицы в adminer (`localhost:8000`)

## 3. Swagger / OpenAPI (swaggo)

- [ ] `go install github.com/swaggo/swag/cmd/swag@latest`
- [ ] `go get github.com/swaggo/gin-swagger github.com/swaggo/files`
- [ ] Общие аннотации в `main.go`: `@title`, `@version`, `@BasePath /api/v1`, `@securityDefinitions.apikey BearerAuth` /
      `@in header` / `@name Authorization`
- [ ] Цель в Makefile: `swag: swag init -g cmd/flyline/main.go -o docs` (+ добавить в `.PHONY`)
- [ ] Роут `GET /swagger/*any` (только в debug-режиме)
- [ ] Открыть `http://localhost:$HTTP_PORT/swagger/index.html`
- [ ] Аннотации на каждом хендлере: `@Summary`, `@Tags`, `@Param`, `@Success`, `@Failure`, `@Router`, у защищённых —
      `@Security BearerAuth`
- [ ] DTO запросов/ответов — отдельные структуры с `json` и `binding` тегами (они попадут в схему)

## 4. Авторизация (refresh token)

Схема:

- **access token**: JWT, живёт 15 мин, отправляется в заголовке `Authorization: Bearer ...`
- **refresh token**: случайная строка (32 байта, `crypto/rand`), живёт 30 дней, лежит в **httpOnly cookie** (`Secure`,
  `SameSite=Strict`, `Path=/api/v1/auth`). В БД хранится только **хэш** (sha256)
- **ротация**: при каждом `/refresh` старый токен отзывается и выдаётся новый
- **повторное использование** уже отозванного токена → отозвать все токены пользователя (вероятно, токен украли)

Зависимости:

- [ ] `go get github.com/golang-jwt/jwt/v5`
- [ ] `golang.org/x/crypto/bcrypt` — хэш паролей

Репозиторий (`features/user`, `features/auth`):

- [ ] `user/repository.go` (таблица `users`): `Create`, `GetByEmail`, `GetByID`
- [ ] `auth/repository.go` (таблица `refresh_tokens`): `Create`, `GetByHash`, `Revoke`, `RevokeAllForUser`

Сервис:

- [ ] `Register(email, password)`: валидация, bcrypt, уникальность email → 409
- [ ] `Login(email, password)` → access + refresh; при неверных данных всегда одна и та же ошибка 401
- [ ] `Refresh(refreshToken)` → проверка хэша / срока / `revoked_at`, ротация, новые access + refresh
- [ ] `Logout(refreshToken)` → отозвать текущий токен
- [ ] `LogoutAll(userID)` → отозвать все токены пользователя

Эндпоинты (`/api/v1/auth`):

- [ ] `POST /register` → 201 + пользователь (без пароля)
- [ ] `POST /login` → `{access_token, expires_in}` + cookie `refresh_token`
- [ ] `POST /refresh` → читает cookie, возвращает новый `access_token` + новую cookie
- [ ] `POST /logout` → отзывает токен, очищает cookie (`MaxAge: -1`)
- [ ] `POST /logout-all` (защищённый)
- [ ] `GET /api/v1/me` (защищённый) → текущий пользователь

Middleware:

- [ ] `AuthRequired`: парсит Bearer, проверяет подпись и `exp`, кладёт `userID` в `gin.Context`
- [ ] Проверять алгоритм подписи (только HS256), иначе 401

Проверка:

- [ ] Весь сценарий в Swagger UI: register → login → me → refresh → logout → refresh (должен вернуть 401)
- [ ] Повторный refresh старым токеном после ротации → 401 + все сессии пользователя отозваны
- [ ] Тесты сервиса auth (`go test ./...`)

## 5. Фронт: типы из OpenAPI

swag v1 генерирует **Swagger 2.0**, а `openapi-typescript` понимает только OpenAPI 3.x. Варианты:

- `npx swagger-typescript-api` — работает со Swagger 2.0 напрямую (типы + клиент)
- или генерировать OpenAPI 3 (swag v2: `swag init --v3.1`), тогда подойдёт `npx openapi-typescript`

- [ ] Выбрать генератор (рекомендация: `openapi-typescript`, если swag v2 со всем справляется, иначе
      `swagger-typescript-api`)
- [ ] Скрипт в `package.json` фронта, например:
      `"gen:api": "openapi-typescript http://localhost:8080/swagger/doc.json -o src/api/schema.d.ts"`
- [ ] Цепочка после изменения API: `make swag` → `npm run gen:api`
- [ ] На фронте: `fetch(..., { credentials: "include" })`, чтобы отправлялась refresh-cookie
- [ ] Интерцептор: на 401 → `POST /auth/refresh` → повторить исходный запрос; если refresh тоже вернул 401 → редирект на
      логин

## Позже (не этап 1)

- [ ] Рейсы, аэропорты, поиск
- [ ] Бронирование и места
- [ ] Оплата
- [ ] Роли (admin / user)
- [ ] Rate limit на `/auth/login`
