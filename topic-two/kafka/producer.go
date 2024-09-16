package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

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

func (kc *KafkaCluster) ProduceMessage(topicName, key string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		color.Red("Error in serializing data to jsondata")
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topicName,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(jsonData),
	}

	partition, offset, err := kc.Producer.SendMessage(msg)
	if err != nil {
		color.Red("Error in SendMessage")
		return err
	}

	color.Cyan("Message sent to %d partition at %d offset successfully to %s topic.", partition, offset, topicName)

	return nil
}

func (kc *KafkaCluster) SendDummyMessages(topicName string, sleepTime int, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	i := 0
	var key string
	for {
		select {
		case <-ctx.Done():
			color.Red("Stopping sending dummy messages to %v topic.", topicName)
			return
		default:
			msg := KafkaMessage{
				Data: fmt.Sprintf("Message#%d", i),
				Time: int(time.Now().UnixMicro()),
			}

			if i%2 == 0 {
				key = "even"
			} else {
				key = "odd"
			}

			err := kc.ProduceMessage(topicName, key, msg)
			if err != nil {
				color.Red("%v", err)
				continue
			}
			i++
			time.Sleep(time.Duration(sleepTime) * time.Second)
		}
	}
}
