package kafka

import (
	"context"
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
	for message := range claim.Messages() {
		log.Printf("Consumed message: Key = %s at Partition = %d, Offset = %d of topic %s",
			string(message.Key), message.Partition, message.Offset, message.Topic)
		log.Printf("%v", message.Value)

		session.MarkMessage(message, "Processed!")
	}

	return nil
}

func (kc *KafkaCluster) CreateConsumer() error {
	config := sarama.NewConfig()
	config.Version = kc.version
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Group.Session.Timeout = 10 * 1000

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

	for {
		if err := kc.Consumer.Consume(ctx, []string{topicName}, handler); err != nil {
			log.Printf("ERROR: %v", err)
		}
	}
}
