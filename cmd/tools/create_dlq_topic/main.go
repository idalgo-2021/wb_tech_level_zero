/////////////////////////////////////
//
// Утилита для создания DLQ-топика
//
/////////////////////////////////////

package main

import (
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func CreateTopic(brokers []string, topic string, partitions int, replicationFactor int) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(10 * time.Second))

	return conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     partitions,
		ReplicationFactor: replicationFactor,
	})
}

func main() {
	brokers := []string{"localhost:9092"}
	topic := "orders-dlq"

	err := CreateTopic(brokers, topic, 1, 1)
	if err != nil {
		log.Fatalf("Failed to create topic: %v", err)
	}

	log.Println("Topic created successfully")
}
