# ZPIF Analyzer

[![CI/CD Pipeline](https://github.com/AkaZver/zpif-analyzer/actions/workflows/ci-cd.yml/badge.svg)](https://github.com/AkaZver/zpif-analyzer/actions/workflows/ci-cd.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=AkaZver_zpif-analyzer&metric=alert_status)](https://sonarcloud.io/summary/overall?id=AkaZver_zpif-analyzer)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=AkaZver_zpif-analyzer&metric=security_rating)](https://sonarcloud.io/summary/overall?id=AkaZver_zpif-analyzer)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=AkaZver_zpif-analyzer&metric=reliability_rating)](https://sonarcloud.io/summary/overall?id=AkaZver_zpif-analyzer)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=AkaZver_zpif-analyzer&metric=sqale_rating)](https://sonarcloud.io/summary/overall?id=AkaZver_zpif-analyzer)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=AkaZver_zpif-analyzer&metric=vulnerabilities)](https://sonarcloud.io/summary/overall?id=AkaZver_zpif-analyzer)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=AkaZver_zpif-analyzer&metric=coverage)](https://sonarcloud.io/summary/overall?id=AkaZver_zpif-analyzer)

Сервис для анализа закрытых паевых инвестиционных фондов (ЗПИФ) недвижимости с использованием LLM.

## Возможности

- **Сравнение фондов**: Таблица с ключевыми метриками (цена пая, NAV, дисконт, Cap Rate, P/NAV, доходность)
- **Детальная информация**: Графики динамики цены и выплат, документы фонда
- **Автоматическая загрузка данных**: Интеграция с MOEX ISS API и investfunds.ru для получения рыночных данных
- **Обогащение через LLM**: Автоматическое заполнение данных фонда при создании через LLM
- **Поиск информации**: Поиск информации о фондах через LLM
- **LLM-анализ**: Анализ документов с извлечением метрик и оценкой рисков
- **Экспорт**: Excel экспорт данных фондов
- **Настройки**: Настройка LLM интеграции

## Технологии

### Backend
- Go 1.26
- Gin (HTTP router)
- GORM (ORM)
- PostgreSQL 16
- JWT аутентификация
- OpenAI API (LLM)
- excelize (Excel)
- goquery (парсинг HTML)
- MOEX ISS API (рыночные данные)
- investfunds.ru (РСП и выплаты)

### Frontend
- React 19 + TypeScript
- Vite
- Tailwind CSS
- Ant Design
- Recharts (графики)
- React Router
- vitest (тестирование)

### Инфраструктура
- Docker & Docker Compose
- Nginx (reverse proxy)
- GitHub Actions (CI/CD)
- SonarCloud (анализ кода)
- Yandex Cloud (хостинг)

## Быстрый старт

### Требования
- Docker и Docker Compose
- API ключ для LLM (задаётся через UI после запуска)

### Запуск

1. Клонируйте репозиторий:
```bash
git clone https://github.com/AkaZver/zpif-analyzer.git
cd zpif-analyzer
```

2. Создайте `.env` файл (опционально):
```bash
cp .env.example .env
# Отредактируйте .env и добавьте API ключи
```

3. Запустите сервисы:
```bash
docker-compose up -d
```

4. Откройте браузер:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080/api/health

### Учётные данные по умолчанию
- **Username**: admin
- **Password**: admin

## Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `DB_HOST` | Хост PostgreSQL | `postgres` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь БД | `zpif` |
| `DB_PASSWORD` | Пароль БД | `zpif` |
| `DB_NAME` | Имя БД | `zpif_analyzer` |
| `DB_SSL_MODE` | SSL режим БД | `disable` |
| `JWT_SECRET` | Секрет для JWT | `change-me-in-production` |
| `SERVER_PORT` | Порт backend | `8080` |

## API Endpoints

### Аутентификация
- `POST /api/auth/login` - Вход
- `GET /api/auth/me` - Текущий пользователь

### Фонды
- `GET /api/funds` - Список фондов
- `GET /api/funds/:id` - Детали фонда
- `POST /api/funds` - Создать фонд
- `PUT /api/funds/:id` - Обновить фонд
- `DELETE /api/funds/:id` - Удалить фонд

### Финансовые данные
- `GET /api/funds/:id/financials` - Финансовые метрики
- `POST /api/funds/:id/financials` - Добавить метрики

### Документы
- `GET /api/funds/:id/documents` - Список документов
- `POST /api/funds/:id/documents` - Загрузить документ
- `DELETE /api/funds/:id/documents/:docId` - Удалить документ
- `GET /api/funds/:id/documents/:docId/download` - Скачать документ
- `POST /api/funds/:id/discover` - Поиск информации через LLM
- `GET /api/funds/:id/discovery-status` - Статус поиска

### Анализ
- `GET /api/funds/:id/analysis` - Последний анализ
- `POST /api/funds/:id/analyze` - Запустить анализ

### LLM настройки
- `GET /api/llm/settings` - Получить настройки
- `PUT /api/llm/settings` - Обновить настройки
- `POST /api/llm/test` - Тест подключения
- `GET /api/llm/models` - Получить список доступных моделей

### Экспорт
- `GET /api/export/excel` - Экспорт в Excel

### Рыночные данные
- `POST /api/funds/:id/fetch-market-data` - Загрузить рыночные данные для фонда (MOEX + investfunds)
- `POST /api/funds/fetch-all-market-data` - Загрузить рыночные данные для всех фондов

## Разработка

### Локальный запуск без Docker

#### Backend
```bash
cd backend
go mod download
go run ./cmd/server
```

#### Frontend
```bash
cd frontend
npm install
npm run dev
```

### Запуск тестов

#### Backend
```bash
cd backend
go test ./... -v
go test ./... -coverprofile=coverage.out
```

#### Frontend
```bash
cd frontend
npm run test
npm run test -- --coverage
```

### Тестирование

Проект придерживается практики обязательного тестирования с целевым покрытием кода **≥ 80%**.

#### Backend
- **Фреймворк**: стандартный `testing` + `testify` + `sqlmock`
- **Парсеры**: тестируются через `httptest.NewServer` (MOEX, investfunds, vsezpif)
- **Сервисы**: dependency injection через интерфейсы, моки через `testify/mock`
- **Handlers**: `httptest` + `gin.TestMode`

#### Frontend
- **Фреймворк**: vitest + @testing-library/react + jsdom
- **API клиент**: mock через `vi.mock('axios')`
- **Компоненты**: рендеринг через `@testing-library/react` с моками контекстов

#### CI/CD
- Тесты запускаются автоматически при каждом push/PR
- Coverage отчёты загружаются в SonarCloud
- Quality Gate требует ≥ 80% покрытия для нового кода

## Структура проекта

```
zpif-analyzer/
├── .github/
│   └── workflows/
│       └── ci-cd.yml          # GitHub Actions workflow
├── backend/
│   ├── cmd/server/            # Точка входа
│   ├── internal/
│   │   ├── config/            # Конфигурация
│   │   ├── models/            # GORM модели
│   │   ├── handlers/          # HTTP handlers
│   │   ├── services/          # Бизнес-логика
│   │   ├── repositories/      # Работа с БД
│   │   ├── parsers/           # Парсеры внешних источников (MOEX, investfunds)
│   │   ├── llm/               # LLM интеграция
│   │   ├── auth/              # JWT аутентификация
│   │   └── middleware/        # CORS middleware
│   ├── migrations/            # SQL миграции
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/        # React компоненты
│   │   ├── pages/             # Страницы
│   │   ├── hooks/             # Custom hooks
│   │   ├── api/               # API клиент
│   │   ├── assets/            # Статические ресурсы (иконки, изображения)
│   │   └── types/             # TypeScript типы
│   ├── nginx.dev.conf         # Nginx конфигурация для разработки
│   ├── nginx.prod.conf        # Nginx конфигурация для production (с SSL)
│   └── Dockerfile
├── docker-compose.yml         # Локальная разработка
├── docker-compose.prod.yml    # Production деплой
├── sonar-project.properties   # SonarCloud конфигурация
├── LICENSE
└── README.md
```

## CI/CD

Проект использует GitHub Actions для автоматизации сборки и деплоя.

### Pipeline

```
┌─────────────────────────────────────────────────────────────┐
│  Trigger: push to master, pull_request to master           │
└─────────────────────────────────────────────────────────────┘
                            │
            ┌───────────────┴───────────────┐
            ▼                               ▼
┌───────────────────────────┐   ┌───────────────────────────┐
│  Build & Test             │   │  SonarCloud Analysis      │
│  - Go build + test        │   │  - Code quality check     │
│  - Frontend test + build  │   │  - Security scan          │
│  - Frontend lint          │   │                           │
└───────────────────────────┘   └───────────────────────────┘
            │
            ▼ (only on push to master)
┌─────────────────────────────────────────────────────────────┐
│  Build & Push Docker Images                                 │
│  - Build backend/frontend images                            │
│  - Push to DockerHub (tags: latest, commit-sha)             │
└─────────────────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────┐
│  Deploy to Yandex Cloud VM                                  │
│  - SSH to VM                                                │
│  - Pull images from DockerHub                               │
│  - docker-compose up -d                                     │
└─────────────────────────────────────────────────────────────┘
```

### Secrets

Для работы CI/CD необходимо настроить следующие secrets в GitHub repository:

| Secret | Описание |
|--------|----------|
| `SONAR_TOKEN` | Токен для SonarCloud |
| `DOCKERHUB_USERNAME` | Имя пользователя DockerHub |
| `DOCKERHUB_TOKEN` | Access token DockerHub |
| `VM_HOST` | IP адрес виртуальной машины |
| `VM_SSH_KEY` | SSH приватный ключ |
| `VM_USER` | Пользователь SSH |

## Развертывание в Yandex Cloud

### Требования

1. Виртуальная машина в Yandex Compute Cloud с Docker и Docker Compose
2. Настроенный SSH доступ
3. Настроенные secrets в GitHub repository

### Первоначальная настройка VM

```bash
# Подключиться к VM по SSH
ssh user@vm-ip

# Создать директорию для проекта
mkdir -p ~/zpif-analyzer
cd ~/zpif-analyzer

# Создать .env файл с production переменными
cat > .env << EOF
DB_PASSWORD=your-secure-password
JWT_SECRET=your-jwt-secret
DOCKERHUB_USERNAME=<your-dockerhub-username>
IMAGE_TAG=latest
EOF
```

### Деплой

Деплой происходит автоматически при push в ветку `master` через GitHub Actions.

Для ручного деплоя:

```bash
# На VM
cd ~/zpif-analyzer
export IMAGE_TAG=latest
export DOCKERHUB_USERNAME=<your-dockerhub-username>
docker-compose -f docker-compose.prod.yml pull
docker-compose -f docker-compose.prod.yml up -d
```

## Интеграция с внешними источниками данных

### MOEX ISS API
Автоматическая загрузка истории котировок (цена пая) с Московской биржи.

**Особенности:**
- Поиск ценной бумаги по ISIN
- Поддержка нескольких board'ов (TQIF, TQBR)
- Fallback для цен: CLOSE → LEGALCLOSEPRICE → WAPRICE
- Пагинация для загрузки полной истории

### investfunds.ru
Автоматическая загрузка РСП (NAV), СЧА и истории выплат.

**Особенности:**
- Парсинг HTML через goquery
- Поле `investfunds_url` в модели Fund для ручной настройки URL фонда
- Извлечение данных из таблиц на странице фонда

### Интерполяция данных
Автоматическое заполнение пропусков в исторических данных РСП методом линейной интерполяции.

**Алгоритм:**
- Если есть предыдущее и следующее значение — линейная интерполяция
- Если есть только предыдущее значение — extrapolation (используется последнее известное)

## Лицензия

Apache License 2.0 - см. файл [LICENSE](LICENSE)

## Авторы

- [AkaZver](https://github.com/AkaZver)
