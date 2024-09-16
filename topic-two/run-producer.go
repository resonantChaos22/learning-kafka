package main

import (
	"log"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/fatih/color"
)

// one producer function to create item
// multiple producer functions to change value of each item
const (
	UPDATE_INTERVAL    = 2
	NET_POSITIVE_DELAY = 5 * time.Minute
)

func (e *Executer) RunProducer() {
	err := e.cluster.CreateProducer()
	if err != nil {
		log.Fatalf("Failed to create Kafka Producer: %v\n", err)
	}
	defer e.wg.Done()
	defer func() {
		if err := e.cluster.Producer.Close(); err != nil {
			log.Fatalf("Failed to close Kafka Producer: %v", err)
		}
		color.Red("Kafka Producer successfully closed")
	}()

	wgProducer := new(sync.WaitGroup)
	wgProducer.Add(2)

	//	create sample messaging here

	go e.ChangeValue(wgProducer, 1)
	go e.ChangeValue(wgProducer, 2)

	<-e.ctx.Done()

	wgProducer.Wait()
}

func (e *Executer) ChangeValue(wg *sync.WaitGroup, itemID int) {
	defer wg.Done()

	item, err := e.store.GetItem(itemID)
	if err != nil {
		log.Fatalf("Error in getting item - %v", err)
		return
	}

	startTime := time.Now()
	netChange := 0.0

	for {
		select {
		case <-e.ctx.Done():
			color.Red("Stopped sending messaages to topic for Item#%d", item.ID)
			return
		default:
			change := randomValueChange()
			netChange += change

			if time.Since(startTime) >= NET_POSITIVE_DELAY {
				if netChange <= 0 {
					change += 1.0 - netChange
					netChange = 1.0
				}
				startTime = time.Now()
			}

			err := e.cluster.ProduceMessage("change_value", item.ID, change)
			if err != nil {
				color.Red("%v", err)
				continue
			}
			time.Sleep(UPDATE_INTERVAL * time.Second)
		}
	}
}

// randomValueChange generates random positive and negative values
func randomValueChange() float64 {
	// Generates a random float between -2.0 and 2.0
	return (rand.Float64() * 4.0) - 2.0
}
