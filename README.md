# ZPIF Analyzer

Сервис для анализа закрытых паевых инвестиционных фондов (ЗПИФ) недвижимости с использованием LLM.

## Возможности

- **Сравнение фондов**: Таблица с ключевыми метриками (цена пая, NAV, дисконт, Cap Rate, P/NAV, доходность)
- **Детальная информация**: Графики динамики цены и выплат, документы фонда
- **Поиск информации**: Поиск информации о фондах через LLM
- **LLM-анализ**: Анализ документов с извлечением метрик и оценкой рисков
- **Экспорт**: Excel экспорт данных фондов
- **Настройки**: Управление фондами, настройка LLM

## Технологии

### Backend
- Go 1.22+
- Gin (HTTP router)
- GORM (ORM)
- PostgreSQL 16
- JWT аутентификация
- OpenAI API (LLM)
- excelize (Excel)

### Frontend
- React 18 + TypeScript
- Vite
- Tailwind CSS
- Ant Design
- Recharts (графики)
- React Router

### Инфраструктура
- Docker & Docker Compose
- Nginx (reverse proxy)

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
```

#### Frontend
```bash
cd frontend
npm run build
```

## Структура проекта

```
zpif-analyzer/
├── backend/
│   ├── cmd/server/          # Точка входа
│   ├── internal/
│   │   ├── config/          # Конфигурация
│   │   ├── models/          # GORM модели
│   │   ├── handlers/        # HTTP handlers
│   │   ├── services/        # Бизнес-логика
│   │   ├── repositories/    # Работа с БД
│   │   ├── llm/             # LLM интеграция
│   │   ├── auth/            # JWT аутентификация
│   │   └── middleware/      # CORS middleware
│   ├── migrations/          # SQL миграции
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── components/      # React компоненты
│   │   ├── pages/           # Страницы
│   │   ├── hooks/           # Custom hooks
│   │   ├── api/             # API клиент
│   │   └── types/           # TypeScript типы
│   ├── nginx.conf
│   └── Dockerfile
├── docker-compose.yml
└── README.md
```

## Лицензия

MIT

## Авторы

- [AkaZver](https://github.com/AkaZver)
