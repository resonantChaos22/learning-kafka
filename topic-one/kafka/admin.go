package kafka

import (
	"fmt"
	"log"
	"strings"

	"github.com/IBM/sarama"
)

type KafkaCluster struct {
	brokers  []string
	version  sarama.KafkaVersion
	Admin    sarama.ClusterAdmin
	Producer sarama.SyncProducer
}

func NewKafkaCluster(brokers []string, version sarama.KafkaVersion) *KafkaCluster {

	return &KafkaCluster{
		brokers: brokers,
		version: version,
	}
}

func (kc *KafkaCluster) CreateAdmin() error {
	config := sarama.NewConfig()
	config.Version = kc.version

	admin, err := sarama.NewClusterAdmin(kc.brokers, config)
	if err != nil {
		return err
	}

	kc.Admin = admin
	log.Println("Kafka ClusterAdmin successfully created")

	return nil
}

func (kc *KafkaCluster) CreateTopic(topicName string, numPartitions, replicationFactor int) error {
	topicDetail := &sarama.TopicDetail{
		NumPartitions:     int32(numPartitions),
		ReplicationFactor: int16(replicationFactor),
		ConfigEntries:     kc.createDefaultConfigEntries(),
	}

	err := kc.Admin.CreateTopic(topicName, topicDetail, false)
	if err != nil {
		return err
	}

	log.Printf("Topic with name %s, number of partitions %d and replication factor %d successfully created.\n", topicName, numPartitions, replicationFactor)
	return nil
}

func (kc *KafkaCluster) ListTopics() error {
	topics, err := kc.Admin.ListTopics()
	if err != nil {
		return err
	}

	builder := new(strings.Builder)
	numTopics := 0
	for topic := range topics {
		if !strings.HasPrefix(topic, "_") {
			builder.WriteString(fmt.Sprintf("%s ", topic))
			numTopics++
		}
	}
	if numTopics == 0 {
		return fmt.Errorf("no topics found")
	}
	log.Printf("Kafka Topics(%d):=\n", numTopics)
	log.Println(builder.String())
	return nil
}

func (kc *KafkaCluster) DeleteTopic(topicName string) error {
	err := kc.Admin.DeleteTopic(topicName)
	if err != nil {
		return err
	}

	log.Printf("Topic with name %s deleted successfully", topicName)
	return nil
}

func (kc *KafkaCluster) createDefaultConfigEntries() map[string]*string {
	return map[string]*string{
		"retention.ms": getStringPtr("60000"),
	}
}

func getStringPtr(s string) *string {
	return &s
}
