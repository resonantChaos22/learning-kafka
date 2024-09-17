package main

import (
	"log"
	"sync"
)

func (e *Executer) RunConsumer() {
	log.Println("Started Running Consumer")
	defer e.wg.Done()

	wgConsumer := new(sync.WaitGroup)
	wgConsumer.Add(2)

	//	run multiple consumers here and how they deal with the code
	go e.cluster.ListenForMessagesFromSingleTopic(VALUE_CHANGE_TOPIC, e.store, wgConsumer, e.ctx)
	go e.cluster.ListenForMessagesFromSingleTopic(DEBEZIUM_ITEM_TOPIC, e.store, wgConsumer, e.ctx)

	<-e.ctx.Done()

	wgConsumer.Wait()
}
