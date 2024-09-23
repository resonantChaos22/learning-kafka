package kafka

import (
	"context"
	"fmt"
	"log"
	"sync"
	"topic-two/items"
	"topic-two/kafka/handlers"

	"github.com/IBM/sarama"
	"github.com/fatih/color"
)

// ChangeValueHandler handles the messages from "value_change" topic
// ItemUpdateHandler handles the messages in `debezium.public.items` topics coming from debezium

// Function to create consumer
func (kc *KafkaCluster) CreateConsumer(groupName ...string) (sarama.ConsumerGroup, error) {
	config := sarama.NewConfig()
	config.Version = kc.version
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	// config.Consumer.Group.Session.Timeout = time.Millisecond * 30
	// config.Consumer.Group.Heartbeat.Interval = time.Millisecond * 2
	group := "Default_Consumer"
	if len(groupName) != 0 {
		group = groupName[0]
	}
	consumerGroup, err := sarama.NewConsumerGroup(kc.brokers, group, config)
	if err != nil {
		return nil, err
	}
	if group == "Default_Consumer" {
		kc.Consumer = consumerGroup
	}
	color.Yellow("%s created!", group)
	return consumerGroup, nil
}

func (kc *KafkaCluster) ListenForValueChangeMessages(store items.Storage, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	groupName := "Value_Change_Consumer"
	topicName := "value_change"
	currCG, err := kc.CreateConsumer(groupName)
	if err != nil {
		log.Printf("Unable to create consumer group for %s due to error: %v", topicName, err)
		return
	}
	defer func() {
		if err := currCG.Close(); err != nil {
			log.Fatalf("Failed to close %s: %v", groupName, err)
			return
		}
		color.Red("%s successfully closed.", groupName)
	}()

	handler := handlers.ChangeValueHandler{
		Store: store,
	}

	color.Green("Listening for messages on %s topic...", topicName)

	for {
		if err := currCG.Consume(ctx, []string{topicName}, handler); err != nil {
			if err.Error() == "context canceled" {
				color.Red("Stoppped listening to %v topic", topicName)
				return
			}
			color.Red("ERROR: %v", err)
		}
	}
}

func (kc *KafkaCluster) ListenForItemChanges(itemID int, itemChan chan<- handlers.DebeziumUpdateMessage, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	groupName := fmt.Sprintf("Item-%d_Change_Consumer", itemID)
	topicName := "debezium.public.items"
	currCG, err := kc.CreateConsumer(groupName)
	if err != nil {
		log.Printf("Unable to create consumer group for %s due to error: %v", topicName, err)
		return
	}
	defer func() {
		if err := currCG.Close(); err != nil {
			log.Fatalf("Failed to close %s: %v", groupName, err)
			return
		}
		color.Red("%s successfully closed.", groupName)
	}()

	handler := handlers.ItemUpdateHandler{
		ID:        itemID,
		ValueChan: itemChan,
	}

	color.Green("Listening for messages on %s topic...", topicName)

	for {
		if err := currCG.Consume(ctx, []string{topicName}, handler); err != nil {
			if err.Error() == "context canceled" {
				color.Red("Stoppped listening to %v topic", topicName)
				return
			}
			color.Red("ERROR: %v", err)
		}
	}
}
