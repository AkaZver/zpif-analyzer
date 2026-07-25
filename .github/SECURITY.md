# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability / Сообщение об уязвимости

**English:**

Please report security vulnerabilities through GitHub's Private Vulnerability Reporting.

To submit a report:
1. Go to the Security tab of this repository
2. Click "Report a vulnerability"
3. Provide detailed information about the issue

You can expect:
- Acknowledgment within 48 hours
- Regular updates on the progress
- Credit in the security advisory (unless you prefer to remain anonymous)

**Русский:**

Пожалуйста, сообщайте об уязвимостях безопасности через GitHub Private Vulnerability Reporting.

Для отправки отчёта:
1. Перейдите на вкладку Security этого репозитория
2. Нажмите "Report a vulnerability"
3. Предоставьте подробную информацию о проблеме

Что ожидать:
- Подтверждение получения в течение 48 часов
- Регулярные обновления о прогрессе
- Упоминание в security advisory (если вы не предпочитаете остаться анонимным)

## Security Best Practices / Лучшие практики безопасности

### Production Deployment

When deploying to production, ensure:

1. **Change default credentials**: Replace default admin password (`admin`) immediately
2. **Secure JWT secret**: Use a strong, random JWT_SECRET (minimum 32 characters)
3. **Database security**: Use strong database passwords and restrict network access
4. **API keys**: Store LLM API keys securely, never commit them to version control
5. **HTTPS**: Enable SSL/TLS in production using `nginx.prod.conf`
6. **Environment variables**: Use `.env` file or secure secret management, never hardcode secrets

### Development

- Never commit `.env` files or API keys to the repository
- Use `sqlmock` and test fixtures instead of real database connections in tests
- Run `npm run lint` and `go vet` before committing code
- Review SonarCloud security hotspots regularly

## Known Security Considerations / Известные соображения безопасности

- **LLM API Keys**: Stored encrypted in database, masked in UI (`****`)
- **JWT Tokens**: Configurable expiration, stored in browser localStorage
- **CORS**: Configured to allow specific origins only
- **Input Validation**: All API endpoints validate and sanitize user input
- **SQL Injection**: Protected by GORM parameterized queries

## Languages / Языки

We accept security reports in both English and Russian.

Мы принимаем отчёты об уязвимостях на английском и русском языках.
