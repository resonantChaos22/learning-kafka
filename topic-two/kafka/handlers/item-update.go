package handlers

import (
	"encoding/json"
	"log"
	"topic-two/items"

	"github.com/IBM/sarama"
)

type ItemUpdateHandler struct {
	ID        int
	ValueChan chan<- DebeziumUpdateMessage
}

type DebeziumUpdateMessage struct {
	Item      items.Item `json:"after"`
	Op        string     `json:"op"`
	TimeStamp int        `json:"ts_ms"`
}

func (ItemUpdateHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (ItemUpdateHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (itemHandler ItemUpdateHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	var msg DebeziumUpdateMessage
	for message := range claim.Messages() {
		err := json.Unmarshal(message.Value, &msg)
		if err != nil {
			log.Printf("Error in unmarshalling - %v", err)
		}
		if msg.Item.ID == itemHandler.ID {
			// color.Cyan("%s", string(message.Key))
			// log.Printf("%v", msg)
			itemHandler.ValueChan <- msg

			session.MarkMessage(message, "Processed!")
		}
	}

	return nil
}
