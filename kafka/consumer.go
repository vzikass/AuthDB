package kafka

import (
	"log"

	"github.com/IBM/sarama"
)

func ConsumeMessage(brokers []string, topic string) error{
	partitionConsumer, err := Consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil{
		return err
	}
	defer partitionConsumer.Close()
	for msg := range partitionConsumer.Messages(){
		log.Printf("Received Message: %s", string(msg.Value))
	}
	return nil
}