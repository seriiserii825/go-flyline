# 03. Единый формат ошибок

Все ошибки API в одном виде, чтобы фронт разбирал их одинаково:

```json
{ "error": { "code": "email_taken", "message": "email already registered" } }
```

`code` нужен для машины (фронт делает `switch` по нему), `message` для человека. Файл
`internal/platform/httpserver/errors.go` уже есть, но пустой.

## Шаги

- [ ] В `errors.go` структуры ответа (они же попадут в Swagger):
  - [ ] `type ErrorBody struct { Code string `json:"code"`; Message string `json:"message"` }`
  - [ ] `type ErrorResponse struct { Error ErrorBody `json:"error"` }`
- [ ] Константы общих кодов: `bad_request`, `validation_failed`, `unauthorized`, `forbidden`, `not_found`,
      `conflict`, `internal`
- [ ] Функция-помощник `func Error(c *gin.Context, status int, code, message string)`, внутри
      `c.AbortWithStatusJSON(status, ErrorResponse{...})`. `Abort`, чтобы дальше по цепочке ничего не выполнилось
- [ ] 500: наружу всегда отдавать общий `message` ("internal server error"), а настоящую ошибку писать в лог
      (`c.Error(err)` или `slog`). Детали БД/стека клиенту не уходят
- [ ] 404 и 405 от самого gin, а не дефолтный текст:
  - [ ] `engine.NoRoute(...)` → `Error(c, 404, "not_found", "route not found")`
  - [ ] `engine.HandleMethodNotAllowed = true` + `engine.NoMethod(...)` → 405
- [ ] Паника → 500 в том же формате: вместо `gin.Recovery()` использовать
      `gin.CustomRecovery(func(c *gin.Context, err any) {...})`
- [ ] (позже, в фичах) каждая фича маппит свои ошибки (`user.ErrEmailTaken` → 409 `email_taken`) в своём
      `handler.go` через `errors.Is` и вызывает этот помощник
- [ ] `go build ./...`

## Проверка

- [ ] `curl -i localhost:8080/api/v1/nope` → `404` + `{"error":{"code":"not_found",...}}`
- [ ] `curl -i -X POST localhost:8080/api/v1/health` → `405` в том же формате
- [ ] (по желанию) временный роут с `panic("x")` → `500` в том же формате + строка в логе, потом удалить
