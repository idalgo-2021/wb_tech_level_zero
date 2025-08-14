
* Чтение из Кафки, валидация, запись в БД.
* Добавить DLQ
* Добавить Кэш
* Добавить свагер



# Создать топик Кафка внутри контейнера
sudo docker exec -it kafka /bin/bash
kafka-topics.sh --create --topic orders --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1



# 
go build -o bin/main ./cmd/main
go build -o bin/order-producer ./cmd/order-producer

// Запуск генератора
go run cmd/order-producer/main.go

// Проверка 
docker exec -it kafka kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic orders --from-beginning --max-messages 5

// создание топика для DLQ
go run ./cmd/tools/create_dlq_topic





# Архитектура
Приложение представляет сервис(модульный монолит), обеспечивающий следующую функциональность:
* 

# Технологии
* Golang
    - zap logger(от Uber)
    - HTTP роутер mux(от gorilla) 
    - горутины
    - gracefull shutdown
* PostgreSQL
    - раскидываем данные по нескольким таблицам, а не сохраняем исходное сообщение в виде jsonb-поля.
    - таблицы БД -orders, deliveries, payments, items. Отдельной таблицы для номенклатуры нет.

* Kafka
    - настраиваемое число консьюмеров
    - коммит при успешной обработке сообщения
    - Retry
    - DLQ
    - есть генератор входящих запросов в Кафку 

    Сценарий с авто-коммитом / ручным коммитом и отправкой в Kafka
        (фокус на данных, не событиях.)
    
            Текущий подход (как ты описал)
            Consumer получает сообщение из Kafka.

            Маппишь и валидируешь заказ.

            Пишешь заказ в БД.

            Отправляешь уведомление (допустим, в другой топик Kafka) о том, что заказ был записан.

            Если всё ок → возвращаешь nil → Kafka-коммит.

            Если отправка уведомления в Kafka не удалась → при следующем чтении проверяешь, есть ли заказ в БД, и, если он там есть, возвращаешь nil → Kafka-коммит.

            Валидация вх.сообщения - "github.com/go-playground/validator/v10"


* Swagger
* Redis
    * кэш
    * прогрев кэша
* Docker(Docker-compose)


# Структура каталогов и файлов проекта


