package kafka

import "github.com/IBM/sarama"


// Creating Topic with code
func CreateTopic(brokers []string, topic string, partitions int32, replicationFactor int16) error {
	config := sarama.NewConfig()
	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil{
		return err
	}
	defer admin.Close()

	topicDetail := &sarama.TopicDetail{
		NumPartitions: partitions,
		ReplicationFactor: replicationFactor,
	}
	err = admin.CreateTopic(topic, topicDetail, false)
	if err != nil{
		return err
	}
	return nil
}