Subscriptions Aggregation
Сервис для управления подписками, их учёта и аналитики с использованием Go, PostgreSQL, ClickHouse, Kafka и GRPC.

Описание
Данный проект позволяет создавать, читать, обновлять и удалять подписки (CRUD).
Поддерживает создание аналитики по суммарной стоимости подписок.
Использует миграции для поддержки схемы базы данных, а также автогенерируемую документацию API через Swagger UI.

Функционал
Управление подписками: создание, получение списка, обновление, удаление

Аггрегация и суммирование стоимости подписок

HTTP API на базе chi router

Логирование запросов с уникальными request ID

Автоматический запуск миграций базы данных через goose

Интеграция со Swagger UI для интерактивной документации API

Требования
Go 1.18+

PostgreSQL с расширением uuid-ossp

Docker (для запуска миграций или контейнеров)

(опционально) ClickHouse, Kafka для аналитики

Установка и запуск
Склонируйте репозиторий:

bash
git clone https://github.com/yourusername/subscribe_aggregation-main.git
cd subscribe_aggregation-main
Настройте параметры подключения к базе в файле конфигурации (укажите параметры по вашему шаблону или через переменные среды).

Запустите миграции (автоматически при старте сервера или вручную через goose):

bash
goose -dir internal/storage/migrations postgres "connection_string" up
Соберите и запустите сервер:

bash
go build -o subscribe_agg ./cmd/main.go
./subscribe_agg
Откройте Swagger UI для изучения API:

text
http://localhost:8080/swagger/index.html
Структура проекта
/cmd/main.go — точка входа, инициализация и запуск сервера

/internal/api — HTTP обработчики API

/internal/config — конфигурация и подключение к базе

/internal/storage — слой доступа к данным, миграции

/pkg/logging — пакет для централизованного логирования

/docs — swagger документы

Пример использования API
Создать подписку: POST /subscriptions

Получить список подписок: GET /subscriptions

Получить подписку по ID: GET /subscriptions/{id}

Обновить подписку: PUT /subscriptions/{id}

Удалить подписку: DELETE /subscriptions/{id}

Получить сумму стоимости: GET /subscriptions/sum

Лицензия
MIT

Контакты
Автор: Андрей
Email: constrictor74@mail.ru
GitHub: https://github.com/AndyBer-creator