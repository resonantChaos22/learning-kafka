package main

import (
	"log"
	"os"
	"strconv"
	"sync"
)

func (e *Executer) RunConsumer() {
	log.Println("Started Running Consumer")
	defer e.wg.Done()

	wgConsumer := new(sync.WaitGroup)
	wgConsumer.Add(2)

	//	run multiple consumers here and how they deal with the code
	go e.cluster.ListenForMessagesFromSingleTopic(VALUE_CHANGE_TOPIC, e.store, wgConsumer, e.ctx)
	if os.Args[4] != "" {
		itemID, err := strconv.Atoi(os.Args[4])
		if err != nil {
			log.Fatalf("%v", err)
		}
		go e.cluster.ListenForMessagesFromSingleTopic(DEBEZIUM_ITEM_TOPIC, e.store, wgConsumer, e.ctx, itemID)
	} else {
		go e.cluster.ListenForMessagesFromSingleTopic(DEBEZIUM_ITEM_TOPIC, e.store, wgConsumer, e.ctx)
	}

	<-e.ctx.Done()

	wgConsumer.Wait()
}
