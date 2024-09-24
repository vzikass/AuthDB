package kafka

import (
	"log"

	"github.com/IBM/sarama"
)

func ConsumeMessage(brokers []string, topic string) error {
	// Start consuming messages from partition 0 of the given topic, starting from the newest message.
	partitionConsumer, err := Consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return err
	}
	// Ensure the partition consumer is properly closed when done.
	defer partitionConsumer.Close()

	// Continuously receive and process messages from the partition.
	for msg := range partitionConsumer.Messages() {
		// Log the value of the received message.
		log.Printf("Received Message: %s", string(msg.Value))
	}
	return nil
}
