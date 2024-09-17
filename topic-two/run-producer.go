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
	wgProducer.Add(3)

	//	create sample messaging here

	go e.ChangeValue(wgProducer, 1, 4.0, 2)
	go e.ChangeValue(wgProducer, 2, 8.0, 4)
	go e.ChangeValue(wgProducer, 3, 20.0, 6)

	<-e.ctx.Done()

	wgProducer.Wait()
}

func (e *Executer) ChangeValue(wg *sync.WaitGroup, itemID int, changeDelta float64, interval int) {
	defer wg.Done()

	// item, err := e.store.GetItem(itemID)
	// if err != nil {
	// 	log.Fatalf("Error in getting item - %v", err)
	// 	return
	// }

	startTime := time.Now()
	netChange := 0.0

	for {
		select {
		case <-e.ctx.Done():
			color.Red("Stopped sending messaages to topic for Item#%d", itemID)
			return
		default:
			change := randomValueChange(float64(changeDelta))
			netChange += change

			if time.Since(startTime) >= NET_POSITIVE_DELAY {
				if netChange <= 0 {
					change += (changeDelta / 4) - netChange
					netChange = changeDelta / 4
				}
				startTime = time.Now()
			}

			err := e.cluster.ProduceMessage(VALUE_CHANGE_TOPIC, itemID, change)
			if err != nil {
				color.Red("%v", err)
				continue
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
}

// randomValueChange generates random positive and negative values
func randomValueChange(changeDelta float64) float64 {
	// Generates a random float between -2.0 and 2.0
	return (rand.Float64() * changeDelta) - (changeDelta / 2)
}
