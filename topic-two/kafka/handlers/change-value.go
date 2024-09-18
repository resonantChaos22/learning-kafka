package handlers

import (
	"log"
	"strconv"
	"topic-two/items"

	"github.com/IBM/sarama"
	"github.com/fatih/color"
)

type ChangeValueHandler struct {
	Store items.Storage
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

		item, err := handler.Store.GetItem(id)
		if err != nil {
			log.Printf("Error in getting the item with id %d: %v", id, err)
			continue
		}
		err = handler.Store.UpdateValue(item.ID, item.Value+delta)
		if err != nil {
			log.Printf("Error in updating the value for item with id %d: %v", id, err)
			continue
		}
		color.Green("Successfully updated %v's value to %f with a delta of %v", item.Name, item.Value+delta, delta)
		session.MarkMessage(message, "Processed!")
	}

	return nil
}
