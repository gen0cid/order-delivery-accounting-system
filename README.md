# Система учета доставки заказов

Полнофункциональная система управления доставкой заказов, построенная на Go с использованием современных технологий: PostgreSQL, Redis, Kafka и Docker.

## 📋 Содержание

- [Описание проекта](#описание-проекта)
- [Архитектура](#архитектура)
- [Технологии](#технологии)
- [Требования](#требования)
- [Быстрый старт](#быстрый-старт)
- [API документация](#api-документация)
- [Конфигурация](#конфигурация)
- [Развертывание](#развертывание)
- [Мониторинг](#мониторинг)
- [Разработка](#разработка)
- [Задачи для доработки](#задачи-для-доработки)

## 🎯 Описание проекта

Система учета доставки заказов - это учебный проект, демонстрирующий лучшие практики разработки на Go. Система позволяет:

- **Управлять заказами**: создание, отслеживание статусов, обновление
- **Управлять курьерами**: регистрация, назначение заказов, отслеживание местоположения
- **Обрабатывать события**: асинхронная обработка через Kafka
- **Кешировать данные**: быстрый доступ через Redis
- **Мониторить состояние**: health checks и метрики

## 🏗 Архитектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │────│   API Gateway   │────│   Load Balancer │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                 │
                        ┌─────────────────┐
                        │   HTTP Server   │
                        │   (Go std lib)  │
                        └─────────────────┘
                                 │
                ┌────────────────┼────────────────┐
                │                │                │
        ┌───────────────┐ ┌─────────────┐ ┌─────────────┐
        │   Handlers    │ │  Services   │ │ Middleware  │
        └───────────────┘ └─────────────┘ └─────────────┘
                │                │                │
        ┌───────────────┐ ┌─────────────┐ ┌─────────────┐
        │   Kafka       │ │ PostgreSQL  │ │    Redis    │
        │  (Events)     │ │(Primary DB) │ │   (Cache)   │
        └───────────────┘ └─────────────┘ └─────────────┘
```

### Компоненты системы:

1. **HTTP API**: RESTful API без внешних фреймворков
2. **Business Logic**: Сервисы для обработки бизнес-логики
3. **Data Layer**: PostgreSQL для персистентности, Redis для кеширования
4. **Event System**: Kafka для асинхронной обработки событий
5. **Monitoring**: Health checks и логирование

## 🛠 Технологии

- **Язык**: Go 1.21+
- **База данных**: PostgreSQL 15
- **Кеш**: Redis 7
- **Очереди**: Apache Kafka
- **Контейнеризация**: Docker & Docker Compose
- **Логирование**: Structured logging (JSON)

### Основные зависимости:

```go
github.com/IBM/sarama v1.41.2          // Kafka client
github.com/go-redis/redis/v8 v8.11.5   // Redis client
github.com/lib/pq v1.10.9              // PostgreSQL driver
github.com/google/uuid v1.3.1          // UUID generation
github.com/sirupsen/logrus v1.9.3      // Structured logging
```

## 📋 Требования

- **Go**: версия 1.21 или выше
- **Docker**: версия 20.0 или выше
- **Docker Compose**: версия 2.0 или выше
- **Make**: для удобства разработки (опционально)

## 🚀 Быстрый старт

### 1. Клонирование репозитория

```bash
git clone <repository-url>
cd delivery-system
```

### 2. Запуск инфраструктуры

```bash
# Запуск всех сервисов (PostgreSQL, Redis, Kafka, приложение)
docker-compose up -d

# Или запуск только инфраструктуры для локальной разработки
docker-compose up -d postgres redis kafka zookeeper
```

### 3. Запуск приложения локально

```bash
# Установка зависимостей
go mod download

# Запуск приложения
go run cmd/server/main.go
```

### 4. Проверка работоспособности

```bash
# Health check
curl http://localhost:8080/health

# Создание курьера
curl -X POST http://localhost:8080/api/couriers \
  -H "Content-Type: application/json" \
  -d '{"name": "Иван Петров", "phone": "+7(999)123-45-67"}'

# Создание заказа
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "Анна Смирнова",
    "customer_phone": "+7(999)987-65-43",
    "delivery_address": "Москва, ул. Ленина, д. 10, кв. 5",
    "items": [
      {"name": "Пицца Маргарита", "quantity": 1, "price": 500.00},
      {"name": "Кока-кола 0.5л", "quantity": 2, "price": 100.00}
    ]
  }'
```

## 📚 API документация

### Заказы (Orders)

#### Создание заказа
```http
POST /api/orders
Content-Type: application/json

{
  "customer_name": "Имя клиента",
  "customer_phone": "+7(999)123-45-67",
  "delivery_address": "Адрес доставки",
  "current_lat": "", 
  "current_lon": "", 
  "items": [
    {
      "name": "Название товара",
      "quantity": 1,
      "price": 100.50
    }
  ]
}
```

#### Получение заказа
```http
GET /api/orders/{order_id}
```

#### Получение списка заказов
```http
GET /api/orders?status=created&courier_id={uuid}&limit=20&offset=0
```

#### Обновление статуса заказа
```http
PUT /api/orders/{order_id}/status
Content-Type: application/json

{
  "status": "in_delivery",
  "courier_id": "uuid-курьера"
}
```

### Курьеры (Couriers)

#### Создание курьера
```http
POST /api/couriers
Content-Type: application/json

{
  "name": "Имя курьера",
  "phone": "+7(999)123-45-67", 
  "current_lat": "55.8051",
  "current_lon": "37.5158"
}
```

#### Получение курьера
```http
GET /api/couriers/{courier_id}
```

#### Получение списка курьеров
```http
GET /api/couriers?status=available&limit=20&offset=0
```

#### Получение доступных курьеров
```http
GET /api/couriers/available
```

#### Обновление статуса курьера
```http
PUT /api/couriers/{courier_id}/status
Content-Type: application/json

{
  "status": "available",
  "current_lat": 55.7558,
  "current_lon": 37.6176
}
```

#### Назначение заказа курьеру
```http
POST /api/couriers/{courier_id}/assign
Content-Type: application/json

{
  "order_id": "uuid-заказа"
}
```
2е задание 
#### Определение оптимального курьера
```http
POST /api/auto-assign/{order_id}
```
### Статусы

#### Статусы заказов:
- `created` - создан
- `accepted` - принят
- `preparing` - готовится
- `ready` - готов к доставке
- `in_delivery` - в доставке
- `delivered` - доставлен
- `cancelled` - отменен

#### Статусы курьеров:
- `offline` - не в сети
- `available` - доступен
- `busy` - занят

### Health Check

```http
GET /health              # Полная проверка всех компонентов
GET /health/readiness    # Проверка готовности к обработке запросов
GET /health/liveness     # Проверка жизнеспособности приложения
```
### Отзывы
#### Оставить отзыв о курьере
```http
POST /api/CourierReview/{courier_id}
Content-Type: application/json

{
  "rating": рейтинг который вы хотите указать,
  "comment": "Ваш комментарий",
  "order_id": "id заказа"
}
```


## ⚙️ Конфигурация

Конфигурация осуществляется через переменные окружения:

### Сервер
```bash
SERVER_HOST=0.0.0.0          # Хост сервера
SERVER_PORT=8080             # Порт сервера
SERVER_READ_TIMEOUT=10       # Таймаут чтения (сек)
SERVER_WRITE_TIMEOUT=10      # Таймаут записи (сек)
```

### База данных
```bash
DB_HOST=localhost            # Хост PostgreSQL
DB_PORT=5432                # Порт PostgreSQL
DB_USER=delivery_user       # Пользователь БД
DB_PASSWORD=delivery_pass   # Пароль БД
DB_NAME=delivery_system     # Название БД
DB_SSL_MODE=disable         # Режим SSL
```

### Redis
```bash
REDIS_HOST=localhost        # Хост Redis
REDIS_PORT=6379            # Порт Redis
REDIS_PASSWORD=            # Пароль Redis (если есть)
REDIS_DB=0                 # Номер БД Redis
```

### Kafka
```bash
KAFKA_BROKERS=localhost:9092              # Брокеры Kafka
KAFKA_GROUP_ID=delivery-service           # ID группы потребителей
KAFKA_TOPIC_ORDERS=orders                 # Топик для заказов
KAFKA_TOPIC_COURIERS=couriers             # Топик для курьеров
KAFKA_TOPIC_LOCATIONS=locations           # Топик для местоположений
```

### Логирование
```bash
LOG_LEVEL=info             # Уровень логирования (debug, info, warn, error)
LOG_FORMAT=json            # Формат логов (json, text)
LOG_FILE=                  # Файл логов (пустой = stdout)
```

## 🐳 Развертывание

### Локальная разработка

1. **Запуск инфраструктуры**:
```bash
docker-compose up -d postgres redis kafka zookeeper
```

2. **Миграции БД**:
```bash
# Выполняются автоматически при запуске PostgreSQL
# Файлы миграций находятся в ./migrations/
```

3. **Запуск приложения**:
```bash
go run cmd/server/main.go
```

### Production

1. **Полный запуск через Docker Compose**:
```bash
docker-compose up -d
```

2. **Проверка статуса**:
```bash
docker-compose ps
curl http://localhost:8080/health
```

3. **Просмотр логов**:
```bash
docker-compose logs -f delivery-app
```

### Kubernetes (для продакшена)

```yaml
# Пример deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: delivery-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: delivery-system
  template:
    metadata:
      labels:
        app: delivery-system
    spec:
      containers:
      - name: delivery-system
        image: delivery-system:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: "postgres-service"
        - name: REDIS_HOST
          value: "redis-service"
        - name: KAFKA_BROKERS
          value: "kafka-service:9092"
        livenessProbe:
          httpGet:
            path: /health/liveness
            port: 8080
          initialDelaySeconds: 30
        readinessProbe:
          httpGet:
            path: /health/readiness
            port: 8080
          initialDelaySeconds: 5
```

## 📊 Мониторинг

### Health Checks

Система предоставляет несколько эндпоинтов для мониторинга:

- `/health` - полная проверка здоровья всех компонентов
- `/health/readiness` - готовность к обслуживанию запросов
- `/health/liveness` - жизнеспособность приложения

### Логирование

Система использует структурированное логирование в формате JSON:

```json
{
  "level": "info",
  "msg": "Order created successfully",
  "order_id": "123e4567-e89b-12d3-a456-426614174000",
  "customer_name": "Анна Смирнова",
  "total_amount": 700,
  "time": "2024-01-15T10:30:00Z"
}
```

### Метрики (рекомендуемые для добавления)

- Количество созданных заказов
- Среднее время доставки
- Количество активных курьеров
- Производительность API (latency, throughput)

## 👨‍💻 Разработка

### Структура проекта

```
delivery-system/
├── cmd/
│   └── server/           # Точка входа приложения
├── internal/
│   ├── config/          # Конфигурация
│   ├── database/        # Работа с БД
│   ├── handlers/        # HTTP обработчики
│   ├── kafka/           # Kafka producer/consumer
│   ├── logger/          # Логирование
│   ├── models/          # Модели данных
│   ├── redis/           # Redis клиент
│   └── services/        # Бизнес-логика
├── migrations/          # SQL миграции
├── docker/             # Docker файлы
├── docs/               # Документация
├── docker-compose.yml  # Локальная разработка
├── Dockerfile          # Production образ
├── go.mod              # Go модули
└── README.md
```
