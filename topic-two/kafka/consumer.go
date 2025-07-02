package kafka

import (
	"context"
	"fmt"
	"log"
	"slices"
	"sync"
	"topic-two/items"
	"topic-two/kafka/handlers"

	"github.com/IBM/sarama"
	"github.com/fatih/color"
)

// ChangeValueHandler handles the messages from "value_change" topic
// ItemUpdateHandler handles the messages in `debezium.public.items` topics coming from debezium

//	TODO: Add a condition to not call CreateConsumer but just return the ConsumerGroup if the group with the same name already exists

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
	kc.ConsumerGroups = append(kc.ConsumerGroups, group)
	return consumerGroup, nil
}

func (kc *KafkaCluster) ListenForValueChangeMessages(store items.Storage, wg *sync.WaitGroup, ctx context.Context, errChan chan error) {
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
			errChan <- fmt.Errorf("failed to close %s: %v", groupName, err)
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

func (kc *KafkaCluster) ListenForAllItemChanges(broadcast *ItemsBroadcast, wg *sync.WaitGroup, ctx context.Context, errChan chan error) {
	kc.ListenForItemChanges(broadcast, wg, ctx, 0, errChan)
}

func (kc *KafkaCluster) ListenForItemChanges(broadcast *ItemsBroadcast, wg *sync.WaitGroup, ctx context.Context, itemID int, errChan chan error) {
	defer wg.Done()
	groupName := fmt.Sprintf("Item-%d_Change_Consumer", itemID)
	topicName := "debezium.public.items"

	if slices.Contains(kc.ConsumerGroups, groupName) {
		color.Green("Already listening to %s", groupName)
		return
	}

	currCG, err := kc.CreateConsumer(groupName)
	if err != nil {
		log.Printf("Unable to create consumer group for %s due to error: %v", topicName, err)
		return
	}
	defer func() {
		if err := currCG.Close(); err != nil {
			errChan <- fmt.Errorf("failed to close %s: %v", groupName, err)
			return
		}
		color.Red("%s successfully closed.", groupName)
	}()

	valueChan := make(chan handlers.DebeziumUpdateMessage, 10)

	handler := handlers.ItemUpdateHandler{
		ID:        itemID,
		ValueChan: valueChan,
	}

	color.Green("Listening for messages on %s topic...", topicName)

	go func(broadcast *ItemsBroadcast) {
		for msg := range valueChan {
			broadcast.Publish(msg)
		}
	}(broadcast)

	for {
		if err := currCG.Consume(ctx, []string{topicName}, handler); err != nil {
			if err.Error() == "context canceled" {
				color.Red("Stoppped listening to %v topic", topicName)
				close(valueChan)
				return
			}
			color.Red("ERROR: %v", err)
		}
	}
}

type ItemsBroadcast struct {
	subscribers map[int][]chan handlers.DebeziumUpdateMessage
	mu          *sync.RWMutex
}

func (ib *ItemsBroadcast) Register(id int) chan handlers.DebeziumUpdateMessage {
	ib.mu.Lock()
	defer ib.mu.Unlock()
	itemChan := make(chan handlers.DebeziumUpdateMessage, 10)
	ib.subscribers[id] = append(ib.subscribers[id], itemChan)
	return itemChan
}

func (ib *ItemsBroadcast) Unregister(id int, itemChan chan handlers.DebeziumUpdateMessage) {
	ib.mu.Lock()
	defer ib.mu.Unlock()

	channels, found := ib.subscribers[id]
	if !found {
		fmt.Println("No channel with the given ID", id)
		return
	}

	idxToRemove := -1
	for i, ch := range channels {
		if ch == itemChan {
			idxToRemove = i
			break
		}
	}

	if idxToRemove == -1 {
		fmt.Println("No channel found for the given ID", id)
		return
	}

	copy(channels[idxToRemove:], channels[idxToRemove+1:])
	channels = channels[:len(channels)-1]

	ib.subscribers[id] = channels
	close(itemChan)

	if len(ib.subscribers[id]) == 0 {
		delete(ib.subscribers, id)
	}
}

func (ib *ItemsBroadcast) Publish(msg handlers.DebeziumUpdateMessage) {
	ib.mu.RLock()
	defer ib.mu.RUnlock()

	for _, ch := range ib.subscribers[0] {
		select {
		case ch <- msg:

		default:
			color.Red("Warning: Channel for all items is full, skipping message.")
		}
	}

	if specificSubscribers, found := ib.subscribers[msg.Item.ID]; found {
		for _, ch := range specificSubscribers {
			select {
			case ch <- msg:

			default:
				color.Red("Warning: Channel for itmemID %d subscriber is full, skipping message.", msg.Item.ID)
			}
		}
	}
}

func NewItemsBroadcast() *ItemsBroadcast {
	return &ItemsBroadcast{
		subscribers: make(map[int][]chan handlers.DebeziumUpdateMessage),
		mu:          new(sync.RWMutex),
	}
}
