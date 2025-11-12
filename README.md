# order-service
Order Service
Сервис для обработки и хранения данных о заказах с использованием Go, PostgreSQL и Kafka.

Функциональность

Прием заказов через Kafka
Сохранение заказов в PostgreSQL
Кэширование заказов в памяти для быстрого доступа
HTTP API для получения информации о заказах
Веб-интерфейс для просмотра заказов



	order-service/
	├── cmd/
	│   ├── app/
	│   │   └── main.go
	│   └── order-generator/
	│       └── main.go
	├── config/
	│   └── config.go
	├── internal/
	│   ├── cache/
	│   │   └── cache.go
	│   ├── db/
	│   │   ├── database.go
	│   │   └── models.go
	│   ├── handlers/
	│   │   └── handlers.go
	│   ├── kafka/
	│   │    └── consumer.go
	│ 	└── validation/
	│       └── validator.go
	├── migrations/
	│   └── init.sql
	├── test/
	│   └── kafka-test.go
	├── web/static/
	│   └── index.html
	├── .dockerignore
	├── docker-compose.yml
	├── Dockerfile
	├── Dockerfile.generator
	├── go.mod
	├── go.sum
	├── .env.example
	├── model.json
	└── README.md



Быстрый запуск с Docker

Клонируйте репозиторий:


	git clone https://github.com/Mamkinhecker/order-service.git


Запустите сервисы с помощью Docker Compose:

	docker-compose up -d --build

Формат сообщений Kafka

Сообщения должны быть в JSON формате, соответствующем модели Order. Пример сообщения можно найти в файле model.json.

Docker контейнеризация


Проект включает Dockerfile и docker-compose.yml для простого развертывания:
 * Dockerfile: Сборка Go приложения
 * docker-compose.yml: Оркестрация сервисов (приложение, PostgreSQL, Redpanda)

Мониторинг

Приложение логирует ключевые события:

 * Подключение к базе данных
 * Сохранение заказов
 * Обработка сообщений Kafka
 * HTTP запросы
