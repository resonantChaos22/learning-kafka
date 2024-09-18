package main

import (
	"log"
	"os"
	"strconv"
	"sync"
	"topic-two/items"

	"github.com/fatih/color"
)

func (e *Executer) Stream() {
	defer e.wg.Done()
	if len(os.Args) <= 4 {
		color.Red("itemID not provided")
	}
	color.Green("Starting Stream...")

	store, err := items.NewPostgresStore()
	if err != nil {
		log.Fatalf("Error in creating Postgres Store: %v", err)
	}
	e.store = store
	wgStream := new(sync.WaitGroup)
	wgStream.Add(1)

	valueChan := make(chan float64)

	itemID, err := strconv.Atoi(os.Args[4])
	if err != nil {
		log.Fatalf("%v", err)
	}
	go e.cluster.ListenForItemChanges(itemID, valueChan, wgStream, e.ctx)

	for {
		select {
		case <-e.ctx.Done():
			wgStream.Wait()
			close(valueChan)
			return
		case value := <-valueChan:
			log.Println(value)
		}
	}
}
