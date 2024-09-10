package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
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

	log.Println("Kafka Producer successfully created!")
	return nil
}

func (kc *KafkaCluster) ProduceMessage(topicName, key string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error in serializing data to jsondata")
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topicName,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(jsonData),
	}

	partition, offset, err := kc.Producer.SendMessage(msg)
	if err != nil {
		log.Println("Error in SendMessage")
		return err
	}

	log.Printf("Message sent to %d partition at %d offset successfully", partition, offset)

	return nil
}

func (kc *KafkaCluster) SendDummyMessages() {

	i := 0
	var key string
	for {
		if i > 10 {
			return
		}
		msg := KafkaMessage{
			Data: fmt.Sprintf("Message#%d", i),
			Time: int(time.Now().UnixMicro()),
		}

		if i%2 == 0 {
			key = "even"
		} else {
			key = "odd"
		}

		err := kc.ProduceMessage("order_details", key, msg)
		if err != nil {
			log.Println(err)
			continue
		}
		i++
		time.Sleep(5 * time.Second)
	}
}
