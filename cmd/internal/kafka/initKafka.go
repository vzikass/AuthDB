package kafka 

import (
	"log"

	"github.com/IBM/sarama"
)

// Interface for the producer
type ProducerInterface interface {
    SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error)
}
// Interface for the consumer
type ConsumerInterface interface{
	ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error)
}
// Mock struct for PartitionConsumer
type MockPartitionConsumer struct {
	Messages chan *sarama.ConsumerMessage
	Errors   chan *sarama.ConsumerError
}



var (
	Producer ProducerInterface
	Consumer ConsumerInterface
	Brokers = []string{"localhost:9092"}
    Topic   = "authdb-topic"
)


// The struct into which we will record Kafka's message
type Message struct{
	Value []byte
}

func InitKafka() (producer sarama.SyncProducer, consumer sarama.Consumer){
	// Create Topic
	// if err := CreateTopic(Brokers, Topic, 1, 1); err != nil{
	// 	log.Fatalf("Failed to create topic: %v", err)
	// }
	// log.Println("Topic created successfully")

	// Producer initialization
	producercfg := sarama.NewConfig()
	producercfg.Producer.Return.Successes = true
	Producer1, err := sarama.NewSyncProducer([]string{"kafka-1:9092", "kafka-2:9093"}, producercfg)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}
	Producer = Producer1

	// Consumer initialization
	consumercfg := sarama.NewConfig()
	consumercfg.Consumer.Return.Errors = true
	// or instead of consumercfg you can use nil (in this case the default consumer settings will be used)
	Consumer1, err := sarama.NewConsumer([]string{"kafka-1:9092", "kafka-2:9093"}, consumercfg) 
	if err != nil {
		log.Fatalln("Failed to start Sarama consumer:", err)
	}
	Consumer = Consumer1
	return Producer1, Consumer1
}