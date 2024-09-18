package handlers

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

type ItemUpdateHandler struct {
	ID        int
	ValueChan chan<- float64
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
			// color.Cyan("%s", string(message.Key))
			// log.Printf("%v", msg)
			itemHandler.ValueChan <- msg.After.Value

			session.MarkMessage(message, "Processed!")
		}
	}

	return nil
}
