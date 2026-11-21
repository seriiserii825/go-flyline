# 02. CORS для фронта

Фронт живёт на другом origin (например, `http://localhost:5173`), поэтому браузер блокирует запросы к API без
CORS-заголовков. Refresh-токен лежит в cookie, поэтому нужен `AllowCredentials: true`. При нём **нельзя**
`AllowOrigins: ["*"]`, только конкретные origin'ы.

## Шаги

- [ ] `go get github.com/gin-contrib/cors`
- [ ] Новая переменная окружения, например `CORS_ALLOWED_ORIGINS=http://localhost:5173` (через запятую, если
      несколько):
  - [ ] добавить в `.env` и `.env.example`
  - [ ] поле в `config.Config` (`envconfig` умеет `[]string` из значения через запятую)
  - [ ] валидация: не пусто, и нет `*` (с credentials он не работает)
- [ ] Создать `internal/platform/middleware/cors.go`, функция `CORS(origins []string) gin.HandlerFunc`, внутри
      `cors.New(cors.Config{...})`:
  - [ ] `AllowOrigins: origins`
  - [ ] `AllowMethods: GET, POST, PUT, PATCH, DELETE, OPTIONS`
  - [ ] `AllowHeaders: Origin, Content-Type, Authorization`
  - [ ] `AllowCredentials: true`
  - [ ] `MaxAge: 12 * time.Hour`: браузер кэширует preflight и не шлёт OPTIONS перед каждым запросом
- [ ] Подключить в `httpserver.New` через `engine.Use(...)`. Origins передать параметром в `New` (httpserver не
      должен сам читать config)
- [ ] Порядок: CORS ставится **до** роутов, иначе preflight `OPTIONS` получит 404
- [ ] `go build ./...`

## Проверка

- [ ] Preflight с разрешённого origin:
      `curl -i -X OPTIONS localhost:8080/api/v1/health -H "Origin: http://localhost:5173" -H "Access-Control-Request-Method: GET"`
      → `204` + `Access-Control-Allow-Origin: http://localhost:5173` + `Access-Control-Allow-Credentials: true`
- [ ] С чужим origin (`-H "Origin: http://evil.com"`) → заголовка `Access-Control-Allow-Origin` нет
