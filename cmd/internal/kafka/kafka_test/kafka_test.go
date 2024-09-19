package kafkatest

import (
	"AuthDB/cmd/internal/kafka"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/mock"
	"AuthDB/mocks"
)

func TestProducerMessage(t *testing.T) {
	// Create mock producer
	mockProducer := new(mocks.ProducerInterface)

	// Mock the behavior of SendMessage
	mockProducer.On("SendMessage", mock.Anything).Return(int32(0), int64(0), nil)

	// Substitute the global variable Producer for mock-producer
	kafka.Producer = mockProducer

	err := kafka.ProduceMessage([]string{"localhost:9092"}, "authdb-topic", "testmessage")
	if err != nil {
		t.Fatalf("Failed to produce message: %v", err)
	}
	mockProducer.AssertExpectations(t)
}

func TestConsumerMessage(t *testing.T) {
	// Create mock Consumer
	mockConsumer := new(mocks.ConsumerInterface)

	// Create a mock PartitionConsumer
	mockPartitionConsumer := new(mocks.PartitionConsumerInterface)
	// Setup mock behavior for messages channel
	mockMessagesChannel := make(chan *sarama.ConsumerMessage)
	close(mockMessagesChannel)
	mockPartitionConsumer.On("Messages").Return((<-chan *sarama.ConsumerMessage)(mockMessagesChannel))
	mockPartitionConsumer.On("Close").Return(nil)
	// Mock the behavior of ConsumePartition
	mockConsumer.On("ConsumePartition", "authdb_topic", int32(0), sarama.OffsetNewest).Return(mockPartitionConsumer, nil)

	kafka.Consumer = mockConsumer

	err := kafka.ConsumeMessage([]string{"localhost:9092"}, "authdb_topic")
	if err != nil {
		t.Fatalf("Failed to consume message: %v", err)
	}
	mockConsumer.AssertExpectations(t)
}