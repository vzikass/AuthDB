package kafka 

import (
	"log"

	"github.com/IBM/sarama"
)

var (
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
	Brokers = []string{"localhost:9092"}
    Topic   = "authdb-topic"
)

// The struct into which we will record Kafka's message
type Message struct{
	Value []byte
}

func InitKafka() {
	var err error

	// Producer initialization
	producercfg := sarama.NewConfig()
	producercfg.Producer.Return.Successes = true
	Producer, err = sarama.NewSyncProducer([]string{"kafka-1:9092", "kafka-2:9093"}, producercfg)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}

	// Consumer initialization
	consumercfg := sarama.NewConfig()
	consumercfg.Consumer.Return.Errors = true
	// or instead of consumercfg you can use nil (in this case the default consumer settings will be used)
	Consumer, err = sarama.NewConsumer([]string{"kafka-1:9092", "kafka-2:9093"}, consumercfg) 
	if err != nil {
		log.Fatalln("Failed to start Sarama consumer:", err)
	}

	// Create Topic
	// if err := CreateTopic(Brokers, Topic, 1, 1); err != nil{
	// 	log.Fatalf("Failed to create topic: %v", err)
	// }
	// log.Println("Topic created successfully")
}