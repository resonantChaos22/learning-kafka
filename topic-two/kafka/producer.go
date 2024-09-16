package kafka

import (
	"strconv"

	"github.com/IBM/sarama"
	"github.com/fatih/color"
)

type KafkaMessage struct {
	Data string `json:"data"`
	Time int    `json:"time"`
}

func (kc *KafkaCluster) CreateProducer() error {
	config := sarama.NewConfig()
	config.Version = kc.version
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	producer, err := sarama.NewSyncProducer(kc.brokers, config)
	if err != nil {
		return err
	}
	kc.Producer = producer

	color.Green("Kafka Producer successfully created!")
	return nil
}

func (kc *KafkaCluster) ProduceMessage(topicName string, key int, value float64) error {

	msg := &sarama.ProducerMessage{
		Topic: topicName,
		Key:   sarama.StringEncoder(strconv.Itoa(key)),
		Value: sarama.StringEncoder(strconv.FormatFloat(value, 'f', 2, 64)),
	}

	partition, offset, err := kc.Producer.SendMessage(msg)
	if err != nil {
		color.Red("Error in SendMessage")
		return err
	}

	color.Cyan("Message sent to %d partition at %d offset successfully to %s topic.", partition, offset, topicName)

	return nil
}
