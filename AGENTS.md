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
- `internal/parsers/` — парсеры внешних источников данных (MOEX, investfunds)
- `internal/llm/` — LLM интеграция (OpenAI-совместимые API)

**Важно:**
- Миграции автоматические через `db.AutoMigrate()` при старте
- Seed данных (фонды + admin user) выполняется при первом запуске
- LLM настройки (API key, URL, модель) задаются через UI и хранятся в БД (таблица `llm_settings`)
- Все API routes защищены JWT middleware (кроме `/api/auth/login` и `/api/health`)

**Тестирование:**
- Используй `sqlmock` для моков БД
- Тесты в файлах `*_test.go` рядом с кодом
- Запуск: `go test ./internal/services -v`

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

## Особенности

1. **LLM интеграция** — настройки задаются через UI (страница "Настройки"), сохраняются в БД. LLM компоненты читают настройки из БД при каждом вызове.
2. **Excel экспорт** — использует библиотеку excelize
3. **Документы** — хранятся в БД в поле `ExtractedText` (текстовое содержимое)
4. **Health checks** — все сервисы имеют health checks в docker-compose
5. **Рыночные данные** — автоматическая загрузка с MOEX ISS API и investfunds.ru
6. **Обогащение через LLM** — автоматическое заполнение данных фонда при создании
7. **Графики** — цена пая отображается только после начала торгов, вертикальная метка "Начало торгов"
