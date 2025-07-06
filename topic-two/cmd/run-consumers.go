package cmd

import (
	"fmt"
	"log"
	"sync"
	"topic-two/items"
)

func (e *Executer) RunConsumer() {
	defer e.wg.Done()
	log.Println("Started Running Consumer")
	store, err := items.NewPostgresStore()
	if err != nil {
		e.errChan <- fmt.Errorf("error in creating postgres store: %v", err)
		return
	}
	e.store = store

	wgConsumer := new(sync.WaitGroup)
	wgConsumer.Add(1)

	go e.cluster.ListenForValueChangeMessages(e.store, wgConsumer, e.ctx, e.errChan)

	<-e.ctx.Done()

	wgConsumer.Wait()
}
