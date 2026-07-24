# AGENTS.md

## Команды разработки

### Docker (основной способ)
```bash
docker-compose up -d          # Запуск всех сервисов
docker-compose up --build -d  # Пересборка и запуск
docker-compose down           # Остановка
```

### Локальная разработка

**Backend:**
```bash
cd backend
go mod download
go run ./cmd/server           # Запуск сервера
go test ./... -v              # Все тесты
go test ./internal/services -v  # Тесты конкретного пакета
go test ./... -coverprofile=coverage.out  # Тесты с coverage
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev                   # Dev сервер (Vite)
npm run build                 # Production сборка (tsc + vite)
npm run lint                  # Проверка кода (oxlint)
npm run test                  # Запуск тестов (vitest)
npm run test:watch            # Тесты в watch режиме
npm run test -- --coverage    # Тесты с coverage report
```

## Архитектура

### Backend (Go + Gin + GORM)

**Структура:**
- `cmd/server/main.go` — точка входа, инициализация всех компонентов
- `internal/handlers/` — HTTP handlers (Gin)
- `internal/services/` — бизнес-логика
- `internal/repositories/` — работа с БД (GORM)
- `internal/models/` — GORM модели
- `internal/parsers/` — парсеры внешних источников данных (MOEX, investfunds)
- `internal/llm/` — LLM интеграция (OpenAI-совместимые API)

**Важно:**
- Миграции автоматические через `db.AutoMigrate()` при старте
- Seed данных (фонды + admin user) выполняется при первом запуске
- LLM настройки (API key, URL, модель) задаются через UI и хранятся в БД (таблица `llm_settings`)
- Все API routes защищены JWT middleware (кроме `/api/auth/login` и `/api/health`)

**Тестирование:**
- **Обязательно**: каждый новый код должен быть покрыт тестами (целевое покрытие ≥ 80%)
- Используй `sqlmock` для моков БД, `testify/mock` для моков зависимостей
- Тесты в файлах `*_test.go` рядом с кодом
- Парсеры (MOEX, investfunds, vsezpif) тестируются через `httptest.NewServer` с настраиваемым `baseURL`
- Сервисы используют интерфейсы для dependency injection (MoexParserI, InvestfundsParserI, VsezpifParserI, FinancialsRepoI, FundRepoI)
- Запуск: `go test ./... -v`
- Coverage: `go test ./... -coverprofile=coverage.out`

**Интеграция с внешними источниками данных:**
- **MOEX ISS API** — загрузка истории котировок (цена пая)
  - Поддержка нескольких board'ов (TQIF, TQBR)
  - Fallback для цен: CLOSE → LEGALCLOSEPRICE → WAPRICE
  - Автоматический поиск по ISIN
- **investfunds.ru** — загрузка РСП (NAV), СЧА и истории выплат
  - Парсинг HTML через goquery
  - Поле `investfunds_url` в модели Fund для ручной настройки URL
- **Интерполяция** — заполнение пропусков в данных РСП методом линейной интерполяции

### Frontend (React 19 + TypeScript + Vite)

**Структура:**
- `src/pages/` — страницы (Dashboard, FundDetails, Settings, Login)
- `src/components/` — переиспользуемые компоненты
- `src/api/client.ts` — единый API клиент (Axios)
- `src/hooks/` — custom hooks (useAuth)
- `src/assets/` — статические ресурсы (building-icon.svg, hero.png)
- `src/types/` — TypeScript типы
- `nginx.dev.conf` — конфигурация для локальной разработки (используется по умолчанию в Dockerfile)
- `nginx.prod.conf` — конфигурация для production с SSL (используется через command override)

**Важно:**
- Линтер: **oxlint** (не eslint!)
- TypeScript с project references (`tsconfig.app.json`, `tsconfig.node.json`)
- UI библиотека: Ant Design 6
- Стилизация: Tailwind CSS 3
- Графики: Recharts

**Тестирование:**
- **Обязательно**: каждый новый компонент/хук/API метод должен быть покрыт тестами (целевое покрытие ≥ 80%)
- Фреймворк: **vitest** + @testing-library/react + @testing-library/jest-dom
- Конфигурация: `vite.config.ts` (секция `test`)
- Setup файл: `src/test/setup.ts`
- Тесты в файлах `*.test.ts` / `*.test.tsx` рядом с кодом
- Mock HTTP через `vi.mock('axios')`
- Запуск: `npm run test`
- Coverage: `npm run test -- --coverage`

## Переменные окружения

Создай `.env` из `.env.example`:
```bash
cp .env.example .env
```

**Обязательные:**
- `DB_*` — настройки PostgreSQL
- `JWT_SECRET` — секрет для JWT токенов
- `ADMIN_PASSWORD` — пароль для начального admin пользователя (используется при первом запуске для seed данных)

**Важно:** В production обязательно изменить значения по умолчанию на безопасные!

## API

**Базовый URL:** `http://localhost:8080/api`

**Аутентификация:**
- `POST /auth/login` — получить JWT токен
- Все остальные endpoints требуют `Authorization: Bearer <token>`

**Основные endpoints:**
- `/funds` — CRUD фондов
- `/funds/:id/financials` — финансовые метрики
- `/funds/:id/documents` — документы фонда
- `/funds/:id/documents/:docId/download` — скачать документ
- `/funds/:id/discover` — автопоиск документов
- `/funds/:id/analyze` — LLM анализ
- `/funds/:id/fetch-market-data` — загрузить рыночные данные (MOEX + investfunds)
- `/funds/fetch-all-market-data` — загрузить рыночные данные для всех фондов
- `/llm/settings` — настройки LLM
- `/export/excel` — экспорт данных

## Дефолтные учётные данные

- **Username:** admin
- **Password:** admin

## Порты

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432

## CI/CD Pipeline

Проект использует GitHub Actions для автоматизации CI/CD.

### Workflow: `.github/workflows/ci-cd.yml`

**Jobs:**

1. **build-and-test** — сборка и тестирование
   - Go: `go mod download`, `go build`, `go test -coverprofile=coverage.out`
   - Frontend: `npm ci`, `npm run test`, `npm run build`, `npm run lint`

2. **sonarcloud** — анализ кода в SonarCloud
   - Запускается после build-and-test
   - Проверяет Quality Gate
   - Анализирует coverage

3. **build-and-push** — сборка и публикация Docker образов
   - Только для push в master
   - Пушит в DockerHub с тегами: `latest`, `<commit-sha>`

4. **deploy** — деплой на Yandex Cloud VM
   - Только для push в master
   - SSH на VM
   - `docker-compose pull && docker-compose up -d`

### Триггеры

- `push` to `master` — полный pipeline с деплоем
- `pull_request` to `master` — только build, test и sonarcloud

### Локальная проверка перед коммитом

```bash
# Backend тесты с coverage
cd backend && go test ./... -coverprofile=coverage.out

# Frontend тесты
cd frontend && npm run test

# Frontend lint
cd frontend && npm run lint

# Frontend build
cd frontend && npm run build
```

## SonarCloud

Конфигурация: `sonar-project.properties`

**Метрики:**
- Coverage (Go tests + Frontend tests)
- Security Rating
- Reliability Rating
- Maintainability Rating
- Vulnerabilities
- Code Smells

**Исключения:**
- `**/node_modules/**`
- `**/vendor/**`
- `**/dist/**`
- `**/migrations/**`
- `**/cmd/**` (точка входа)
- `**/test/**` (тестовые setup файлы)

## Production деплой

### Файл: `docker-compose.prod.yml`

Использует готовые образы из DockerHub вместо локальной сборки:

```yaml
backend:
  image: ${DOCKERHUB_USERNAME}/zpif-backend:${IMAGE_TAG:-latest}
frontend:
  image: ${DOCKERHUB_USERNAME}/zpif-frontend:${IMAGE_TAG:-latest}
```

### Переменные для production

```bash
DB_PASSWORD=<secure-password>
JWT_SECRET=<secure-secret>
DOCKERHUB_USERNAME=<your-dockerhub-username>
IMAGE_TAG=<commit-sha или latest>
```

### Ручной деплой на VM

```bash
cd ~/zpif-analyzer
export IMAGE_TAG=latest
export DOCKERHUB_USERNAME=<your-dockerhub-username>
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d
```

## Особенности

1. **LLM интеграция** — настройки задаются через UI (страница "Настройки"), сохраняются в БД. LLM компоненты читают настройки из БД при каждом вызове.
2. **Прокси для LLM** — поддержка HTTP/HTTPS прокси для обхода гео-блокировок (например, OpenRouter блокирует РФ). Настройки прокси задаются через UI: URL, логин, пароль. Включается чекбоксом "Использовать прокси".
3. **Excel экспорт** — использует библиотеку excelize
4. **Документы** — хранятся в БД в поле `ExtractedText` (текстовое содержимое)
5. **Health checks** — все сервисы имеют health checks в docker-compose
6. **Рыночные данные** — автоматическая загрузка с MOEX ISS API и investfunds.ru
7. **Обогащение через LLM** — автоматическое заполнение данных фонда при создании
8. **Графики** — цена пая отображается только после начала торгов, вертикальная метка "Начало торгов"
9. **CI/CD** — автоматический деплой при push в master через GitHub Actions
10. **SonarCloud** — автоматический анализ качества кода при каждом PR и push
11. **Security** — Docker контейнеры запускаются от non-root пользователя, секреты хранятся в GitHub Secrets
12. **Тестирование** — обязательное написание тестов для backend и frontend, целевое покрытие ≥ 80%, автоматическая проверка в CI/CD
