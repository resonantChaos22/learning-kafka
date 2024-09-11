package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

// ConsumerGroupHandler handles the actual claiming process
type ConsumerGroupHandler struct{}

func (ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msg := new(KafkaMessage)
	for message := range claim.Messages() {
		log.Printf("Consumed message: Key = %s at Partition = %d, Offset = %d of topic %s",
			string(message.Key), message.Partition, message.Offset, message.Topic)
		json.Unmarshal(message.Value, msg)
		log.Printf("%v", msg)

		session.MarkMessage(message, "Processed!")
	}

	return nil
}

func (kc *KafkaCluster) CreateConsumer() error {
	config := sarama.NewConfig()
	config.Version = kc.version
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	// config.Consumer.Group.Session.Timeout = time.Millisecond * 30
	// config.Consumer.Group.Heartbeat.Interval = time.Millisecond * 2

	group := "OrderCG"
	consumerGroup, err := sarama.NewConsumerGroup(kc.brokers, group, config)
	if err != nil {
		return err
	}
	kc.Consumer = consumerGroup
	return nil
}

func (kc *KafkaCluster) ListenForMessagesFromSingleTopic(topicName string) {
	handler := ConsumerGroupHandler{}
	ctx := context.Background()
	log.Printf("Listening for messages on %s topic...", topicName)

	for {
		if err := kc.Consumer.Consume(ctx, []string{topicName}, handler); err != nil {
			log.Printf("ERROR: %v", err)
		}
	}
}
