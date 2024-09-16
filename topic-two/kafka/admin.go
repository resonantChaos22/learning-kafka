package kafka

import (
	"fmt"
	"strings"

	"github.com/IBM/sarama"
	"github.com/fatih/color"
)

type KafkaCluster struct {
	brokers           []string
	version           sarama.KafkaVersion
	topics            []string
	Admin             sarama.ClusterAdmin
	Producer          sarama.SyncProducer
	Consumer          sarama.ConsumerGroup
	DebeziumConnector string
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
	color.Green("Kafka ClusterAdmin successfully created")

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
	kc.topics = append(kc.topics, topicName)

	color.Green("Topic with name %s, number of partitions %d and replication factor %d successfully created.\n", topicName, numPartitions, replicationFactor)
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
			kc.topics = append(kc.topics, topic)
		}
	}
	if numTopics == 0 {
		return fmt.Errorf("no topics found")
	}
	color.Yellow("Kafka Topics(%d):=\n", numTopics)
	color.Magenta(builder.String())
	return nil
}

func (kc *KafkaCluster) DeleteTopic(topicName string) error {
	err := kc.Admin.DeleteTopic(topicName)
	if err != nil {
		return err
	}

	color.Green("Topic with name %s deleted successfully", topicName)
	return nil
}

func (kc *KafkaCluster) createDefaultConfigEntries() map[string]*string {
	return map[string]*string{
		"retention.ms": getStringPtr("300000"),
	}
}

func getStringPtr(s string) *string {
	return &s
}
