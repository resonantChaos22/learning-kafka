package kafka

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/IBM/sarama"
	"github.com/fatih/color"
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
		color.HiMagenta("Consumed message: Key = %s at Partition = %d, Offset = %d of topic %s",
			string(message.Key), message.Partition, message.Offset, message.Topic)
		json.Unmarshal(message.Value, msg)
		color.HiMagenta("%v", msg)

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

func (kc *KafkaCluster) ListenForMessagesFromSingleTopic(topicName string, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	handler := ConsumerGroupHandler{}
	color.Green("Listening for messages on %s topic...", topicName)

	for {
		if err := kc.Consumer.Consume(ctx, []string{topicName}, handler); err != nil {
			if err.Error() == "context canceled" {
				color.Red("Stoppped listening to %v topic", topicName)
				return
			}
			color.Red("ERROR: %v", err)
		}
	}
}

func (kc *KafkaCluster) ListenForMessagesFromMulitpleTopics(topics []string, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	handler := ConsumerGroupHandler{}
	color.Green("Listening for messages on the given topics...")

	for {
		if err := kc.Consumer.Consume(ctx, topics, handler); err != nil {
			if err.Error() == "context canceled" {
				color.Red("Stoppped listening to given topics")
				return
			}
			color.Red("ERROR: %v", err)
		}
	}
}
