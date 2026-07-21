# ZPIF Analyzer — План реализации

## Обзор проекта

Сервис для сбора, анализа и сравнения российских ЗПИФ недвижимости. Собирает данные из документов фондов (отчёты оценщика, КИД, ПДУ), анализирует через LLM, хранит агрегированные данные в PostgreSQL, отображает сравнительную таблицу на фронте, позволяет экспортировать/импортировать данные в Excel.

## Стек технологий

| Слой       | Технология                               |
|------------|------------------------------------------|
| Frontend   | React 18 + Vite, TypeScript, Tailwind CSS, Ant Design (таблицы), Recharts (графики), Vitest + RTL (тесты) |
| Backend    | Go 1.22+, Gin (HTTP router), GORM (ORM), testify (тесты) |
| Database   | PostgreSQL 16                            |
| LLM        | OpenAI-совместимый API (ChatGPT / YandexGPT через совместимый endpoint) |
| Excel      | excelize (Go) на бэке                    |
| Web Search | SerpAPI / Exa (для поиска документов фондов в интернете) |
| Deploy     | Docker Compose (3 контейнера: frontend+nginx, backend, postgres) |
| Auth       | JWT (логин/пароль)                       |

## Дизайн

Визуальный стиль повторяет **snowball-income.com**:
- Тёмная тема (фон `#1a1a2e` / `#16213e`, карточки `#0f3460` / `#1a1a2e`)
- Акцентный цвет — фиолетовый/пурпурный (`#5d507e`, `#7c5cbf`)
- Шрифты: Inter / system-ui (sans-serif)
- Карточная раскладка с закруглёнными углами и мягкими тенями
- Таблицы — Ant Design с кастомной тёмной темой
- Иконки — Material Icons (Google Fonts)

---

## Список фондов (начальный)

| Название              | ISIN            | Тикер   |
|-----------------------|-----------------|---------|
| Парус ОЗН             | RU000A1022Z1    | —       |
| Акцент 5              | RU000A10DQF7    | —       |
| ВИМ РД                | RU000A102N77    | —       |
| Современная коллекция | RU000A10CQ02    | —       |

---

## Отслеживаемые параметры ЗПИФ

### Основные (general)
- `name` — название фонда
- `isin` — ISIN код
- `ticker` — тикер на бирже
- `management_company` — управляющая компания
- `real_estate_segment` — сегмент недвижимости (склады/офисы/ТЦ/ЦОД/жильё)
- `fund_start_date` — дата старта фонда
- `fund_end_date` — дата завершения фонда
- `qualified_investor_required` — требуется ли статус квалифицированного инвестора
- `number_of_properties` — количество объектов
- `main_tenants` — основные арендаторы
- `has_market_maker` — наличие маркет-мейкера

### Финансовые метрики (financials, с привязкой к дате)
- `unit_price_rub` — цена пая, руб.
- `nav_per_unit_rub` — РСП (расчётная стоимость пая), руб.
- `nav_total_mln_rub` — СЧА (стоимость чистых активов), млн руб.
- `discount_to_nav_pct` — дисконт/премия к РСП, %
- `cap_rate_pct` — Cap Rate, %
- `p_nav` — P/NAV (цена / СЧА на пай)
- `p_affo` — P/AFFO
- `noi_yield_pct` — рентабельность по NOI, %
- `annual_payout_rub` — годовая выплата на пай, руб.
- `payout_yield_pct` — доходность от выплат (до налога), %
- `payout_yield_after_tax_pct` — доходность от выплат (после налога), %
- `total_return_pct` — полная доходность за год (выплаты + тело), %
- `payout_frequency` — периодичность выплат (ежемесячно/ежеквартально/полугодие)
- `debt_to_nav_ratio` — долговая нагрузка к СЧА
- `management_fee_pct` — комиссии УК, %
- `trading_volume_mln_rub` — средний дневной объём торгов, млн руб.
- `payout_stability` — стабильность выплат (оценка 1–5 или low/medium/high)
- `rent_indexation_pct` — индексация аренды, %
- `irr_forecast_pct` — прогноз IRR, %

### LLM-анализ
- `llm_analysis_raw` — сырой ответ LLM
- `llm_analysis_summary` — краткое резюме анализа
- `llm_risk_assessment` — оценка рисков от LLM
- `llm_pros_cons` — плюсы и минусы фонда
- `llm_last_updated` — дата последнего LLM-анализа
- `llm_model_used` — использованная модель

### Документы фонда
- `document_type` — тип документа (оценщик, КИД, ПДУ, отчётность)
- `document_url` — ссылка на документ
- `document_source` — источник: `auto` (LLM нашла в интернете) / `manual` (ручная загрузка)
- `document_source_url` — страница, откуда документ был обнаружен (для auto)
- `document_upload_date` — дата загрузки/обнаружения
- `document_content_hash` — хэш контента (для проверки изменений)
- `document_status` — статус: `pending` / `downloaded` / `analyzed` / `error`

---

## Структура проекта

```
zpif-analyzer/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # точка входа
│   ├── internal/
│   │   ├── config/                  # конфигурация (env vars)
│   │   ├── models/                  # GORM модели
│   │   ├── handlers/                # HTTP handlers (Gin)
│   │   ├── services/                # бизнес-логика
│   │   ├── repositories/            # работа с БД
│   │   ├── llm/                     # LLM-клиент (OpenAI API)
│   │   ├── websearch/               # поиск документов в интернете (SerpAPI/Exa)
│   │   ├── fetcher/                 # скачивание HTML/PDF по URL
│   │   ├── excel/                   # экспорт/импорт Excel
│   │   ├── auth/                    # JWT middleware
│   │   └── middleware/              # CORS, logging, auth
│   ├── migrations/                  # SQL миграции
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── components/              # UI компоненты
│   │   ├── pages/
│   │   │   ├── Dashboard/           # главная с таблицей сравнения
│   │   │   ├── FundDetails/         # детальная страница фонда
│   │   │   ├── Settings/            # настройки (LLM, фонды)
│   │   │   └── Login/               # авторизация
│   │   ├── api/                     # API клиент (axios)
│   │   ├── hooks/                   # кастомные хуки
│   │   ├── types/                   # TypeScript типы
│   │   ├── styles/                  # Tailwind конфиг + глобальные стили
│   │   └── utils/
│   ├── public/
│   ├── index.html
│   ├── vite.config.ts
│   ├── tailwind.config.ts
│   ├── tsconfig.json
│   ├── Dockerfile
│   ├── package.json
│   └── nginx.conf
├── docker-compose.yml
├── .env.example
└── PLAN.md
```

---

## Пошаговый план реализации

### Фаза 1: Инициализация проекта и инфраструктура

- [ ] **1.1** Создать структуру каталогов backend/frontend
- [ ] **1.2** Инициализировать Go-модуль (`go mod init`), установить зависимости: gin, gorm, gorm/postgres, excelize, testify, jwt-go, godotenv, goquery (HTML parsing), unipdf (PDF extraction)
- [ ] **1.3** Инициализировать React+Vite проект: `npm create vite@latest`, установить TypeScript, Tailwind CSS, Ant Design, axios, recharts, vitest, react-testing-library
- [ ] **1.4** Настроить Tailwind с кастомной темой (цвета, шрифты из snowball-income)
- [ ] **1.5** Создать `.env.example` с переменными окружения (DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASS, OPENAI_API_KEY, OPENAI_BASE_URL, WEBSEARCH_API_KEY, WEBSEARCH_PROVIDER, JWT_SECRET)
- [ ] **1.6** Настроить `docker-compose.yml` (postgres:16-alpine, backend, frontend+nginx)
- [ ] **1.7** Написать Dockerfile для backend (multi-stage build: build → scratch/alpine)
- [ ] **1.8** Написать Dockerfile для frontend (node build → nginx)
- [ ] **1.9** Написать nginx.conf для фронтенда (проксирование /api → backend, SPA fallback)

### Фаза 2: База данных

- [ ] **2.1** Спроектировать схему БД:
  - `funds` — основные данные фонда (id, name, isin, ticker, management_company, segment, dates, qualified_investor, properties, tenants, market_maker, created_at, updated_at)
  - `fund_financials` — финансовые метрики (id, fund_id, snapshot_date, все метрики из списка выше, created_at)
  - `fund_documents` — загруженные документы (id, fund_id, type, url, source [auto/manual], source_url, upload_date, content_hash, status [pending/downloaded/analyzed/error])
  - `llm_analyses` — результаты LLM-анализа (id, fund_id, document_id, model_used, raw_response, summary, risk_assessment, pros_cons, created_at)
  - `llm_settings` — настройки LLM (id, api_key_encrypted, base_url, model_name, updated_at)
  - `users` — пользователи (id, username, password_hash, created_at)
- [ ] **2.2** Создать GORM-модели в `internal/models/`
- [ ] **2.3** Написать AutoMigrate в точке входа (или использовать golang-migrate для SQL-миграций)
- [ ] **2.4** Настроить подключение к PostgreSQL с пулом коннектов
- [ ] **2.5** Сидирование начальных фондов (4 фонда из списка)

### Фаза 3: Backend — CRUD API

- [ ] **3.1** Реализовать repository-слой (`fund_repository.go`) — CRUD для funds, financials, documents, llm_analyses
- [ ] **3.2** Реализовать service-слой (`fund_service.go`) — бизнес-логика, валидация
- [ ] **3.3** Реализовать handler-слой (`fund_handler.go`):
  - `GET /api/funds` — список фондов с последними финансовыми данными
  - `GET /api/funds/:id` — детали фонда + история метрик
  - `POST /api/funds` — добавить фонд
  - `PUT /api/funds/:id` — обновить фонд
  - `DELETE /api/funds/:id` — удалить фонд
  - `GET /api/funds/:id/financials` — финансовые данные (с фильтром по датам)
  - `POST /api/funds/:id/financials` — добавить срез финансовых данных
  - `GET /api/funds/comparison?ids=1,2,3` — сравнение нескольких фондов
- [ ] **3.4** Настроить Gin router с middleware (CORS, logger, recovery)
- [ ] **3.5** Написать unit-тесты для repositories (с mock БД или testcontainers)
- [ ] **3.6** Написать unit-тесты для services
- [ ] **3.7** Написать unit-тесты для handlers (httptest + mock services)
- [ ] **3.8** Покрытие ≥80%

### Фаза 4: Backend — Excel экспорт/импорт

- [ ] **4.1** Реализовать `excel/exporter.go` — формирование Excel-файла с данными фондов:
  - Лист «Фонды» — общая информация
  - Лист «Финансы» — последние финансовые метрики
  - Лист «Анализ» — результаты LLM-анализа
  - Форматирование, заголовки, типы данных
- [ ] **4.2** Реализовать `excel/importer.go` — парсинг Excel-файла и восстановление данных в БД:
  - Валидация структуры файла
  - Маппинг колонок на поля моделей
  - Upsert-логика (обновление существующих, создание новых)
- [ ] **4.3** Handler:
  - `GET /api/export/excel` — скачать Excel со всеми данными
  - `POST /api/import/excel` — загрузить Excel и восстановить данные
- [ ] **4.4** Unit-тесты для экспорта (проверка структуры и содержимого)
- [ ] **4.5** Unit-тесты для импорта (различные сценарии: корректный файл, повреждённый, пустой)
- [ ] **4.6** Покрытие ≥80%

### Фаза 5: Backend — LLM-интеграция и автопоиск документов

**5A. Поиск и скачивание документов (авто-режим)**

- [ ] **5.1** Реализовать `websearch/client.go` — клиент для поиска в интернете:
  - Поддержка SerpAPI и/или Exa API
  - Поиск по запросу: `"{название фонда}" OR "{ISIN}" документы отчёт оценщика КИД`
  - Парсинг результатов: URL, заголовок, сниппет
  - Фильтрация результатов (отсев нерелевантных)
- [ ] **5.2** Реализовать `fetcher/fetcher.go` — скачивание контента по URL:
  - HTTP GET с таймаутами, следование редиректам
  - Скачивание HTML-страниц (для парсинга ссылок на документы)
  - Скачивание PDF/документов в локальное хранилище (volume)
  - Извлечение текста из PDF (использовать pdfcpu или аналог)
- [ ] **5.3** Реализовать `llm/discoverer.go` — LLM-агент для обнаружения документов:
  - Шаг 1: web search по ISIN + названию фонда → список URL-кандидатов
  - Шаг 2: fetch HTML каждой страницы-кандидата
  - Шаг 3: отправить HTML в LLM с промптом «найди ссылки на документы фонда (отчёт оценщика, КИД, ПДУ, отчётность)»
  - Шаг 4: LLM возвращает структурированный список `{url, type, title}`
  - Шаг 5: скачать документы по найденным URL
  - Шаг 6: сохранить в `fund_documents` с `source=auto`
  - Дедупликация по content_hash (не скачивать уже имеющиеся)

**5B. LLM-анализ документов**

- [ ] **5.4** Реализовать `llm/client.go` — OpenAI-совместимый клиент:
  - Отправка промпта + контекста (текст документа)
  - Парсинг ответа (структурированный JSON через function calling)
  - Поддержка кастомного base_url (для YandexGPT и др.)
  - Таймауты и retry-логика
- [ ] **5.5** Разработать промпт-шаблоны для анализа:
  - Извлечение финансовых метрик из отчёта оценщика
  - Анализ КИД (ключевые инвестиционные документы)
  - Общая оценка рисков и плюсов/минусов фонда
- [ ] **5.6** Реализовать `llm/analyzer.go` — сервис анализа:
  - Принимает текст/PDF документа → отправляет в LLM → парсит структурированный ответ
  - Сохраняет результат в `llm_analyses`
  - Обновляет финансовые метрики фонда из извлечённых данных
  - Отмечает документ как `status=analyzed`

**5C. Ручная загрузка документов (fallback)**

- [ ] **5.7** Handler для ручной загрузки:
  - `POST /api/funds/:id/documents` — multipart upload файла
  - Сохранение в `fund_documents` с `source=manual`
  - Автоматический запуск анализа после загрузки

**5D. API endpoints**

- [ ] **5.8** Handler:
  - `POST /api/funds/:id/discover` — запустить автопоиск документов для фонда
  - `POST /api/funds/discover-all` — запустить автопоиск для всех фондов
  - `GET /api/funds/:id/discovery-status` — статус последнего поиска (найдено/скачано/ошибок)
  - `POST /api/funds/:id/analyze` — запустить LLM-анализ (по имеющимся документам)
  - `GET /api/funds/:id/analysis` — получить последний анализ
  - `GET /api/llm/settings` — получить настройки LLM
  - `PUT /api/llm/settings` — обновить настройки LLM (api_key, base_url, model, websearch_provider)
  - `POST /api/funds/:id/documents` — ручная загрузка документа
  - `GET /api/funds/:id/documents` — список документов (с указанием source: auto/manual)

**5E. Тесты**

- [ ] **5.9** Unit-тесты для websearch-клиента (mock HTTP)
- [ ] **5.10** Unit-тесты для fetcher (mock HTTP)
- [ ] **5.11** Unit-тесты для LLM-клиента (mock HTTP)
- [ ] **5.12** Unit-тесты для discoverer (mock websearch + fetcher + LLM)
- [ ] **5.13** Unit-тесты для analyzer (mock LLM)
- [ ] **5.14** Покрытие ≥80%

### Фаза 6: Backend — Аутентификация

- [ ] **6.1** Реализовать `auth/jwt.go` — генерация и валидация JWT-токенов
- [ ] **6.2** Реализовать `auth/handler.go`:
  - `POST /api/auth/login` — логин (username + password → JWT)
  - `POST /api/auth/register` — регистрация (опционально, можно отключить)
  - `GET /api/auth/me` — текущий пользователь
- [ ] **6.3** JWT middleware для защиты API-эндпоинтов
- [ ] **6.4** Хеширование паролей (bcrypt)
- [ ] **6.5** Seed начального пользователя (admin/admin) при первом запуске
- [ ] **6.6** Unit-тесты для auth (генерация токенов, middleware, хеширование)

### Фаза 7: Frontend — Базовый каркас

- [ ] **7.1** Настроить Vite + React + TypeScript
- [ ] **7.2** Настроить Tailwind CSS с кастомной темой:
  ```
  colors: {
    primary: { DEFAULT: '#7c5cbf', light: '#9b7ed8', dark: '#5d507e' },
    surface: { DEFAULT: '#1a1a2e', light: '#16213e', card: '#0f3460' },
    accent: { DEFAULT: '#e94560', light: '#ff6b6b' },
    text: { primary: '#e0e0e0', secondary: '#a0a0a0', muted: '#6c6c6c' }
  }
  fontFamily: { sans: ['Inter', 'system-ui', 'sans-serif'] }
  ```
- [ ] **7.3** Создать Layout-компонент (sidebar + header + content area) — стиль как у snowball-income
- [ ] **7.4** Настроить роутинг (React Router): `/`, `/funds/:id`, `/settings`, `/login`
- [ ] **7.5** Настроить API-клиент (axios с interceptor для JWT)
- [ ] **7.6** Создать контекст авторизации (AuthContext)
- [ ] **7.7** Настроить Ant Design с тёмной темой (ConfigProvider)

### Фаза 8: Frontend — Страница сравнения (Dashboard)

- [ ] **8.1** Создать компонент `ComparisonTable` на основе Ant Design Table:
  - Колонки: Название, УК, Цена пая, РСП, Дисконт, Cap Rate, P/NAV, Доходность выплат, Полная доходность, Долг/СЧА, Квал, и др.
  - Сортировка по любой колонке
  - Фильтрация по УК, сегменту, квал/неквал
  - Цветовая индикация (зелёный = хорошо, красный = плохо)
  - Горизонтальная прокрутка на мобильных
- [ ] **8.2** Создать компонент `FundCard` — мини-карточка фонда (для мобильной версии)
- [ ] **8.3** Создать хук `useFunds` для загрузки данных с API
- [ ] **8.4** Панель фильтров (сегмент, УК, наличие квал-статуса)
- [ ] **8.5** Кнопка «Обновить данные» — триггер автопоиска + LLM-анализа для всех фондов
- [ ] **8.6** Кнопка «Экспорт в Excel» — скачивание файла
- [ ] **8.7** Кнопка «Импорт из Excel» — загрузка файла
- [ ] **8.8** Unit-тесты для таблицы, фильтров, карточек

### Фаза 9: Frontend — Детальная страница фонда

- [ ] **9.1** Компонент `FundDetails`:
  - Шапка: название, ISIN, тикер, УК, сегмент
  - Карточки с ключевыми метриками (цена, РСП, дисконт, доходность)
  - График истории цены пая и РСП (Recharts)
  - График истории выплат
  - Секция LLM-анализа (summary, риски, плюсы/минусы)
- [ ] **9.2** Секция документов фонда:
  - Список документов с указанием источника (авто/ручная) и статуса
  - Кнопка «Найти документы в интернете» — запуск автопоиска (discover)
  - Кнопка «Загрузить вручную» — upload для закрытых данных
  - Индикатор прогресса поиска/скачивания
  - Возможность удалить/перезагрузить отдельный документ
- [ ] **9.3** Кнопка «Запустить анализ» — вызов LLM по всем документам фонда
- [ ] **9.4** Unit-тесты для компонентов

### Фаза 10: Frontend — Страница настроек

- [ ] **10.1** Секция «Управление фондами»:
  - Таблица фондов (название, ISIN, тикер, УК)
  - Кнопка «Добавить фонд» → модальное окно с формой
  - Кнопки «Редактировать» / «Удалить» для каждого фонда
  - Валидация ISIN (формат)
- [ ] **10.2** Секция «Настройки LLM и поиска»:
  - Поле API Key (password input)
  - Поле Base URL (для OpenAI-совместимых API)
  - Поле Model Name (dropdown: gpt-4o, gpt-4o-mini, yandexgpt и т.д.)
  - Поле Web Search Provider (dropdown: SerpAPI, Exa)
  - Поле Web Search API Key
  - Кнопка «Тест подключения к LLM» — отправка тестового запроса
  - Кнопка «Тест поиска» — тестовый web search запрос
  - Кнопка «Сохранить»
- [ ] **10.3** Секция «Данные»:
  - Кнопка «Экспорт всех данных в Excel»
  - Drag&drop зона для импорта Excel
- [ ] **10.4** Unit-тесты для форм и валидации

### Фаза 11: Frontend — Авторизация

- [ ] **11.1** Страница логина (в стиле snowball-income):
  - Поля username + password
  - Кнопка «Войти»
  - Обработка ошибок
- [ ] **11.2** Защищённые роуты (redirect на /login если нет токена)
- [ ] **11.3** Хранение JWT в localStorage, авто-refresh
- [ ] **11.4** Unit-тесты для auth flow

### Фаза 12: Интеграционное тестирование и отладка

- [ ] **12.1** End-to-end тесты основных сценариев (ручное тестирование):
  - Запуск docker-compose, проверка доступности всех сервисов
  - Логин → Dashboard → просмотр таблицы → переход к фонду
  - Настройки → добавление фонда → возврат на Dashboard
  - Экспорт/импорт Excel
  - LLM-анализ (с реальным API ключом)
- [ ] **12.2** Проверка покрытия тестами ≥80% на бэке и фронте
- [ ] **12.3** Исправление найденных багов

### Фаза 13: Подготовка к деплою

- [ ] **13.1** Финализация docker-compose.yml:
  - Health checks для всех контейнеров
  - Volumes для postgres (персистентность) и для загруженных документов
  - Переменные окружения через .env
- [ ] **13.2** Создать `.env.example` с описанием всех переменных
- [ ] **13.3** README.md с инструкциями по запуску (локально + docker)
- [ ] **13.4** Проверка сборки образов и запуска с нуля

### Фаза 14: Деплой в Yandex Cloud (опционально, после основного тестирования)

- [ ] **14.1** Создать VM в Yandex Cloud (Compute Cloud)
- [ ] **14.2** Установить Docker + Docker Compose на VM
- [ ] **14.3** Настроить security group (порты 80, 443)
- [ ] **14.4** Задеплоить через docker-compose (или SSH + pull + up)
- [ ] **14.5** Настроить домен + HTTPS (Let's Encrypt / Yandex Certificate Manager)

---

## Приоритеты и порядок выполнения

| Приоритет | Фазы       | Описание                              |
|-----------|------------|---------------------------------------|
| P0        | 1–2        | Инициализация проекта, БД, структура  |
| P0        | 3          | CRUD API для фондов                   |
| P1        | 7–8        | Фронт: каркас + таблица сравнения     |
| P1        | 5          | LLM + автопоиск документов            |
| P1        | 9–10       | Фронт: детали фонда + настройки       |
| P2        | 4          | Excel экспорт/импорт                  |
| P2        | 6, 11      | Авторизация                           |
| P3        | 12–14      | Тестирование, деплой                  |

---

## Зависимости между фазами

```
1 → 2 → 3 → 4 (Excel)
         ↓
         5A (web search + fetcher) → 5B (LLM анализ) → 9 (детали фонда)
         ↓
         6 (Auth) → 11 (фронт auth)

7 (каркас фронта) → 8 (таблица) → 9 → 10 → 11
```

---

## Критерии приёмки

1. `docker-compose up` запускает все 3 контейнера, приложение доступно на `localhost:3000`
2. На Dashboard отображается таблица сравнения с 4 начальными фондами
3. В настройках можно добавить/удалить/отредактировать фонд
4. В настройках можно указать LLM API-ключ и Web Search API-ключ, проверить подключение
5. Кнопка «Найти документы» на странице фонда запускает автопоиск — LLM ищет документы по ISIN/названию в интернете, скачивает их
6. Ручная загрузка документов работает (для закрытых данных)
7. LLM-анализ запускается по найденным/загруженным документам и результаты отображаются на странице фонда
8. Экспорт в Excel скачивает файл со всеми данными
9. Импорт из Excel восстанавливает данные в БД
10. Покрытие unit-тестами ≥80% на бэке и фронте
11. Авторизация работает (логин/пароль → JWT → защищённые роуты)
