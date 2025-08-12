
* Чтение из Кафки, валидация, запись в БД.
* Добавить Кэш
* Добавить свагер



# Создать топик Кафка внутри контейнера
sudo docker exec -it kafka /bin/bash
kafka-topics.sh --create --topic orders --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1



# 
go build -o bin/main ./cmd/main
go build -o bin/order-producer ./cmd/order-producer

// Запуск
go run cmd/order-producer/main.go

// Проверка 
docker exec -it kafka kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic orders --from-beginning --max-messages 5


