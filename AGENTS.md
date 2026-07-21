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
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev                   # Dev сервер (Vite)
npm run build                 # Production сборка (tsc + vite)
npm run lint                  # Проверка кода (oxlint)
```

## Архитектура

### Backend (Go + Gin + GORM)

**Структура:**
- `cmd/server/main.go` — точка входа, инициализация всех компонентов
- `internal/handlers/` — HTTP handlers (Gin)
- `internal/services/` — бизнес-логика
- `internal/repositories/` — работа с БД (GORM)
- `internal/models/` — GORM модели
- `internal/llm/` — LLM интеграция (OpenAI API)
- `internal/websearch/` — клиенты для поиска (SerpAPI, Exa)
- `internal/fetcher/` — загрузка документов

**Важно:**
- Миграции автоматические через `db.AutoMigrate()` при старте
- Seed данных (фонды + admin user) выполняется при первом запуске
- LLM функциональность опциональна — работает только если `OPENAI_API_KEY` задан
- Все API routes защищены JWT middleware (кроме `/api/auth/login` и `/api/health`)

**Тестирование:**
- Используй `sqlmock` для моков БД
- Тесты в файлах `*_test.go` рядом с кодом
- Запуск: `go test ./internal/services -v`

### Frontend (React 19 + TypeScript + Vite)

**Структура:**
- `src/pages/` — страницы (Dashboard, FundDetails, Settings, Login)
- `src/components/` — переиспользуемые компоненты
- `src/api/client.ts` — единый API клиент (Axios)
- `src/hooks/` — custom hooks (useAuth)
- `src/types/` — TypeScript типы

**Важно:**
- Линтер: **oxlint** (не eslint!)
- TypeScript с project references (`tsconfig.app.json`, `tsconfig.node.json`)
- UI библиотека: Ant Design 6
- Стилизация: Tailwind CSS 3
- Графики: Recharts

## Переменные окружения

Создай `.env` из `.env.example`:
```bash
cp .env.example .env
```

**Обязательные:**
- `DB_*` — настройки PostgreSQL
- `JWT_SECRET` — секрет для JWT токенов

**Опциональные:**
- `OPENAI_API_KEY` — включает LLM анализ и автопоиск документов
- `OPENAI_BASE_URL` — для совместимых API (по умолчанию OpenAI)
- `WEBSEARCH_API_KEY` — для автопоиска документов (SerpAPI или Exa)

## API

**Базовый URL:** `http://localhost:8080/api`

**Аутентификация:**
- `POST /auth/login` — получить JWT токен
- Все остальные endpoints требуют `Authorization: Bearer <token>`

**Основные endpoints:**
- `/funds` — CRUD фондов
- `/funds/:id/financials` — финансовые метрики
- `/funds/:id/documents` — документы фонда
- `/funds/:id/discover` — автопоиск документов
- `/funds/:id/analyze` — LLM анализ
- `/llm/settings` — настройки LLM
- `/export/excel` — экспорт данных

## Дефолтные учётные данные

- **Username:** admin
- **Password:** admin

## Порты

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432

## Особенности

1. **LLM интеграция** — если `OPENAI_API_KEY` не задан, функции анализа и автопоиска недоступны
2. **Web Search** — поддерживает SerpAPI и Exa (настраивается через `WEBSEARCH_PROVIDER`)
3. **Excel экспорт** — использует библиотеку excelize
4. **Документы** — сохраняются в Docker volume `backend_documents:/app/documents`
5. **Health checks** — все сервисы имеют health checks в docker-compose
