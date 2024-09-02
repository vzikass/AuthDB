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

type Message struct{
	Value []byte
}

func InitKafka() {
	var err error

	// Producer initialization
	producercfg := sarama.NewConfig()
	producercfg.Producer.Return.Successes = true
	Producer, err = sarama.NewSyncProducer([]string{"kafka:9092"}, producercfg)
	if err != nil {
		log.Fatalln("Failed to start Sarama producer:", err)
	}

	// Consumer initialization
	consumercfg := sarama.NewConfig()
	consumercfg.Consumer.Return.Errors = true
	// or instead of consumercfg you can use nil (in this case the default consumer settings will be used)
	Consumer, err = sarama.NewConsumer([]string{"kafka:9092"}, consumercfg) 
	if err != nil {
		log.Fatalln("Failed to start Sarama consumer:", err)
	}
}
