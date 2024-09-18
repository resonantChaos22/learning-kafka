package main

import (
	"log"
	"sync"
	"topic-two/items"
)

func (e *Executer) RunConsumer() {
	defer e.wg.Done()
	log.Println("Started Running Consumer")
	store, err := items.NewPostgresStore()
	if err != nil {
		log.Fatalf("Error in creating Postgres Store: %v", err)
	}
	e.store = store

	wgConsumer := new(sync.WaitGroup)
	wgConsumer.Add(1)

	go e.cluster.ListenForValueChangeMessages(e.store, wgConsumer, e.ctx)

	<-e.ctx.Done()

	wgConsumer.Wait()
}
