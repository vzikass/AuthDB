package kafka

import (
	"log"

	"github.com/IBM/sarama"
)

// ProduceMessage sends a message to a Kafka topic and logs its partition and offset.
// It connects to the Kafka cluster specified by brokers and publishes to the given topic.
func ProduceMessage(brokers []string, topic, message string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	partition, offset, err := Producer.SendMessage(msg)
	if err != nil {
		return err
	}

	// Log the successful message storage with topic, partition, and offset details.
	// Example log: web | 2024/09/12 16:52:12 Message is stored in topic(authdb-topic)/partition(0)/offset(5)
	log.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)
	return nil
}
