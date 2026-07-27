# AGENTS.md

## Команды разработки

### Docker (основной способ)
```bash
docker-compose up -d          # Запуск всех сервисов
docker-compose up --build -d  # Пересборка и запуск
docker-compose down           # Остановка
```

**ВАЖНО:** После любых изменений файлов frontend или backend обязательно запускай `docker-compose up --build -d` для пересборки и перезапуска контейнеров.

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
  - `fund_handler.go` — CRUD фондов, документы, анализ
  - `auth_handler.go` — аутентификация, JWT middleware
  - `llm_handler.go` — настройки LLM, тест подключения
  - `excel_handler.go` — экспорт в Excel
  - `market_data_handler.go` — загрузка рыночных данных
- `internal/services/` — бизнес-логика
  - `fund_service.go` — управление фондами, документами
  - `auth_service.go` — аутентификация пользователей
  - `llm_service.go` — работа с LLM настройками
  - `excel_service.go` — экспорт данных в Excel с форматированием
  - `market_data_service.go` — загрузка и обработка рыночных данных
- `internal/repositories/` — работа с БД (GORM)
- `internal/models/` — GORM модели
- `internal/parsers/` — парсеры внешних источников данных
  - `moex_parser.go` — MOEX ISS API (котировки)
  - `investfunds_parser.go` — investfunds.ru (РСП, СЧА, выплаты)
  - `vsezpif_parser.go` — vsezpif.ru (комиссии, объекты, сегменты)
- `internal/llm/` — LLM интеграция (OpenAI-совместимые API)
  - `client.go` — HTTP клиент для LLM API, функции `ExtractJSON` и `SanitizeJSON` для обработки ответов
  - `discoverer.go` — поиск документов через LLM
  - `analyzer.go` — анализ документов и извлечение метрик
  - `prompts.go` — промпты для LLM

**Важно:**
- Миграции автоматические через `db.AutoMigrate()` при старте
- Seed данных (фонды + admin user) выполняется при первом запуске
- LLM настройки (API key, URL, модель) задаются через UI и хранятся в БД (таблица `llm_settings`)
- Все API routes защищены JWT middleware (кроме `/api/auth/login` и `/api/health`)
- **Hard delete**: все модели используют физическое удаление (без soft delete), при удалении фонда каскадно удаляются все связанные данные
- **Порядок удаления**: из-за foreign key constraints удаление выполняется в порядке: analyses → documents → financials → fund
- **Фильтрация будущих дат**: все запросы к финансовым данным автоматически фильтруют записи с `snapshot_date > NOW()`
- **LLM ответы**: для надёжного парсинга используются `ExtractJSON` (извлечение JSON из markdown, умных кавычек, BOM) и `SanitizeJSON` (исправление неэкранированных кавычек внутри строк)
- **Таймауты**: HTTP-клиент LLM — 180 секунд, контекст для поиска — 180 секунд
- **Часовой пояс**: имена файлов документов создаются с московским временем (Europe/Moscow, GMT+3)

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
- **vsezpif.ru** — загрузка дополнительных данных фонда
  - Комиссия УК, количество объектов, сегмент недвижимости
  - Поиск по ISIN через JSON API
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
- `/funds-with-financials` — список фондов с последними финансовыми данными (оптимизированный для Dashboard, устраняет N+1 проблему)
- `/funds/:id/financials` — финансовые метрики
- `/funds/:id/documents` — документы фонда
- `/funds/:id/documents/:docId/download` — скачать документ
- `/funds/:id/discover` — автопоиск документов
- `/funds/:id/analyze` — LLM анализ (без обновления финансовых данных)
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
3. **Обработка LLM ответов** — функции `ExtractJSON` и `SanitizeJSON` для надёжного парсинга ответов от различных LLM моделей:
   - Извлечение JSON из markdown code blocks
   - Удаление BOM (byte order mark)
   - Замена умных двойных кавычек (`""`) на обычные двойные кавычки
   - Замена умных одинарных кавычек (`''`) на обычные одинарные кавычки
   - Исправление неэкранированных кавычек внутри строковых значений
   - Поддержка JSON с одинарными кавычками как разделителями
4. **Excel экспорт** — использует библиотеку excelize с форматированием:
   - Один лист "Фонды" с 15 колонками (Название, ISIN, Тикер, Квал, УК, Сегмент, Цена пая, РСП, Дисконт к РСП, Cap Rate, СЧА, P/NAV, P/AFFO, Доходность выплат, Комиссия УК)
   - Рубли: `#,##0.00" "[$₽-419]` (пробелы между тысячами, запятая как десятичный, символ рубля в конце)
   - Проценты: `[$-419]0.00%` (русская локаль, запятая как десятичный)
   - Числа: `[$-419]0.00` (P/NAV, P/AFFO с запятой как десятичным)
5. **Документы** — хранятся в БД в поле `ExtractedText` (текстовое содержимое)
6. **Health checks** — все сервисы имеют health checks в docker-compose
7. **Рыночные данные** — автоматическая загрузка с MOEX ISS API, investfunds.ru и vsezpif.ru
8. **Обогащение через LLM** — автоматическое заполнение данных фонда при создании
9. **Графики** — цена пая отображается только после начала торгов, вертикальная метка "Начало торгов"
10. **Часовой пояс** — имена файлов документов создаются с московским временем (Europe/Moscow, GMT+3)
11. **CI/CD** — автоматический деплой при push в master через GitHub Actions
12. **SonarCloud** — автоматический анализ качества кода при каждом PR и push
13. **Security** — Docker контейнеры запускаются от non-root пользователя, секреты хранятся в GitHub Secrets
14. **Тестирование** — обязательное написание тестов для backend и frontend, целевое покрытие ≥ 80%, автоматическая проверка в CI/CD
15. **Фильтрация будущих дат** — автоматическая фильтрация данных с датами в будущем на всех уровнях (парсеры, сервисы, репозитории) для предотвращения некорректных данных на графиках
16. **Hard delete** — физическое удаление записей без soft delete, каскадное удаление связанных данных при удалении фонда
17. **Оптимизация производительности** — endpoint `/funds-with-financials` для Dashboard устраняет проблему N+1 запросов, параллельная загрузка данных на FundDetails
