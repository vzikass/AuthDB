// The creation of the topic is done in the docker compose file
package kafka

import (
	"log"

	"github.com/IBM/sarama"
)

// Creating Topic with code
func CreateTopicIfNotExist(brokers []string, topic string) error {
	admin, err := sarama.NewClusterAdmin(brokers, sarama.NewConfig())
	if err != nil {
		return err
	}
	defer admin.Close()
	topics, err := admin.ListTopics()
	if err != nil {
		return err
	}

	// Check if the topic exists
	if _, exist := topics[topic]; !exist {
		// Create the topic if it doesn't exist
		topicDetail := &sarama.TopicDetail{
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
		err = admin.CreateTopic(topic, topicDetail, false)
		if err != nil {
			return err
		}
		log.Printf("Created topic %s\n", topic)
	}
	return nil
}
