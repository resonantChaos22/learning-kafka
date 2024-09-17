package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"topic-two/items"

	"github.com/IBM/sarama"
	"github.com/fatih/color"
)

// ChangeValueHandler handles the messages from "value_change" topic
type ChangeValueHandler struct {
	store items.Storage
}

type DebeziumUpdateMessage struct {
	After DebeziumAfter `json:"after"`
	Op    string        `json:"op"`
}

type DebeziumAfter struct {
	Id    int     `json:"id"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

func (ChangeValueHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (ChangeValueHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (handler ChangeValueHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		id, err := strconv.Atoi(string(message.Key))
		if err != nil {
			log.Printf("Error in getting id: %v", err)
			continue
		}
		delta, err := strconv.ParseFloat(string(message.Value), 64)
		if err != nil {
			log.Printf("Error in getting delta: %v", err)
			continue
		}

		item, err := handler.store.GetItem(id)
		if err != nil {
			log.Printf("Error in getting the item with id %d: %v", id, err)
			continue
		}
		err = handler.store.UpdateValue(item.ID, item.Value+delta)
		if err != nil {
			log.Printf("Error in updating the value for item with id %d: %v", id, err)
			continue
		}
		color.Green("Successfully updated %v's value to %f with a delta of %v", item.Name, item.Value+delta, delta)
		session.MarkMessage(message, "Processed!")
	}

	return nil
}

// ItemUpdateHandler handles the messages in `debezium.public.items` topics coming from debezium
type ItemUpdateHandler struct {
	ID int
}

func (ItemUpdateHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (ItemUpdateHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (itemHandler ItemUpdateHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msg := new(DebeziumUpdateMessage)
	for message := range claim.Messages() {
		err := json.Unmarshal(message.Value, msg)
		if err != nil {
			log.Printf("Error in unmarshalling - %v", err)
		}
		if msg.After.Id == itemHandler.ID {
			color.Cyan("%s", string(message.Key))
			log.Printf("%v", msg)

			session.MarkMessage(message, "Processed!")
		}
	}

	return nil
}

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

func (kc *KafkaCluster) ListenForMessagesFromSingleTopic(topicName string, store items.Storage, wg *sync.WaitGroup, ctx context.Context, id ...int) {
	defer wg.Done()
	itemID := 0
	if len(id) != 0 {
		itemID = id[0]
	}
	log.Printf("HELLO=%d", itemID)
	currConsumerGroup, err := kc.CreateConsumer(fmt.Sprintf("%s_%d_Consumer", topicName, itemID))
	log.Println("Created Consumer")
	if err != nil {
		log.Printf("Unable to create consumer group due to error: %v", err)
		return
	}
	defer func() {
		if err := currConsumerGroup.Close(); err != nil {
			log.Fatalf("Failed to close Kafka Consumer Group: %v", err)
		}
		color.Red("Kafka Consumer Group successfully closed")
	}()
	var handler sarama.ConsumerGroupHandler
	switch topicName {
	case "value_change":
		handler = ChangeValueHandler{
			store: store,
		}
	case "debezium.public.items":
		handler = ItemUpdateHandler{
			ID: itemID,
		}
		log.Printf("%v", handler)
	default:
		handler = ChangeValueHandler{
			store: store,
		}
	}
	color.Green("Listening for messages on %s topic...", topicName)

	for {
		if err := currConsumerGroup.Consume(ctx, []string{topicName}, handler); err != nil {
			if err.Error() == "context canceled" {
				color.Red("Stoppped listening to %v topic", topicName)
				return
			}
			color.Red("ERROR: %v", err)
		}
	}
}
