package kafkatest

import (
	"AuthDB/cmd/internal/kafka"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/mock"
)

// Mock for Sarama SyncProducer
type MockSyncProducer struct {
	mock.Mock
}

type MockConsumer struct {
	mock.Mock
}

func (m *MockSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	args := m.Called(msg)
	return args.Get(0).(int32), args.Get(1).(int64), args.Error(2)
}

func TestProducerMessage(t *testing.T) {
	// Create mock producer
	mockProducer := new(MockSyncProducer)

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

func (m *MockConsumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	args := m.Called(topic, partition, offset)
	return args.Get(0).(sarama.PartitionConsumer), args.Error(1)
}

func TestConsumerMessage(t *testing.T) {
	// Create mock Consumer
	mockConsumer := new(MockConsumer)

	// Mock the behavior of ConsumePartition
	mockPartitionConsumer := new(kafka.MockPartitionConsumer)
	mockConsumer.On("ConsumePartition", "authdb_topic", int32(0), sarama.OffsetNewest).Return(mockPartitionConsumer, nil)

	kafka.Consumer = mockConsumer

	err := kafka.ConsumeMessage([]string{"localhost:9092"}, "authdb-topic")
	if err != nil{
		t.Fatalf("Failed to consume message: %v", err)
	}
	mockConsumer.AssertExpectations(t)
}
