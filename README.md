# DNS Resolver API

Микросервис для хранения и обновления соответствий между FQDN и IP-адресами.

## 📌 Основные возможности
- Добавление FQDN для мониторинга
POST /api/fqdns
Content-Type: application/json

{
  "fqdn": "example.com"
}

- Автоматическое обновление IP-адресов (каждые 5 минут)

- Поиск всех FQDN по IP
GET /api/fqdns?ip=8.8.8.8

- Поиск всех IP по FQDN
GET /api/ips?fqdn=example.com

### Технологии
- Язык: Go 1.23
- Фреймворк: Echo
- База данных: PostgreSQL 15


### Запуск
1. docker-compose up 

### Остановка
1. docker-compose down

