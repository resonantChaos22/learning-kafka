package cmd

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"topic-two/items"
	"topic-two/kafka"

	"github.com/fatih/color"
)

const (
	DEBEZIUM_CONNECT_URL = "http://localhost:8083/connectors"
	VALUE_CHANGE_TOPIC   = "value_change"
	DEBEZIUM_ITEM_TOPIC  = "debezium.public.items"
)

type Executer struct {
	cluster *kafka.KafkaCluster
	wg      *sync.WaitGroup
	ctx     context.Context
	store   items.Storage
	errChan chan error
}

func (e *Executer) SetWg(wg *sync.WaitGroup) {
	e.wg = wg
	e.wg.Add(1)
}

func (e *Executer) Wait() {
	e.wg.Wait()
}

func (e *Executer) SetupKafka() {
	err := e.cluster.CreateAdmin()

	if err != nil {
		e.errChan <- fmt.Errorf("failed to create kafka cluster admin: %v", err)
	}

	defer func() {
		if err := e.cluster.Admin.Close(); err != nil {
			e.errChan <- fmt.Errorf("failed to closekafka cluster admin: %v", err)
		}
		color.Red("Kafka ClusterAdmin successfully closed!")
	}()

	err = e.cluster.CreateTopic(VALUE_CHANGE_TOPIC, 3, 2)
	if err != nil {
		e.errChan <- fmt.Errorf("failed to create the topic: %v", err)
	}

	func() {
		startTime := time.Now() // Capture the start time
		for {
			if time.Since(startTime) > 30*time.Second {
				color.Red("Unable to list topics. There's some issue")
				return
			}
			checkTime := time.Now()
			err := e.cluster.ListTopics()
			color.Magenta("Took %dms to try to list topics", time.Since(checkTime).Milliseconds())
			if err == nil {
				elapsed := time.Since(startTime) // Calculate elapsed time
				color.Cyan("It took %d milliseconds to achieve sync", elapsed.Milliseconds())
				return
			}
			time.Sleep(2 * time.Millisecond)
		}
	}()
}

func (e *Executer) SetupDebezium() {
	isConnectorPresent, err := e.cluster.CheckDebeziumConnector(DEBEZIUM_CONNECT_URL, "pg_connector")
	if err != nil {
		e.errChan <- fmt.Errorf("error in checking presence of connector - %v", err)
	}

	if !isConnectorPresent {
		err = e.cluster.CreateDebeziumConnector(DEBEZIUM_CONNECT_URL)
		if err != nil {
			e.errChan <- fmt.Errorf("error in creating the connector - %v", err)
		}
	}

	log.Println("Debezium Connector Is Active!")
}

func (e *Executer) SetupDB() {
	store, err := items.NewPostgresStore()
	if err != nil {
		e.errChan <- fmt.Errorf("error in creating postgres store: %v", err)
	}
	e.store = store

	err = e.store.CreateItemTable()
	if err != nil {
		e.errChan <- fmt.Errorf("error in creating item table: %v", err)
	}

	items := []*items.Item{{Name: "ItemA", Value: 500.0}, {Name: "ItemB", Value: 550.0}, {Name: "ItemC", Value: 450.0}}

	for _, item := range items {
		err = e.store.CreateItem(item)
		if err != nil {
			e.errChan <- fmt.Errorf("error in adding item to table: %v", err)
		}
	}
}

func (e *Executer) GetErrors() chan error {
	return e.errChan
}

func (e *Executer) Setup() {
	e.SetupKafka()
	e.SetupDB()
	e.SetupDebezium()
}

func NewExecuter(cluster *kafka.KafkaCluster, wg *sync.WaitGroup, ctx context.Context, store items.Storage) *Executer {
	return &Executer{
		cluster: cluster,
		wg:      wg,
		ctx:     ctx,
		store:   store,
		errChan: make(chan error, 10),
	}
}
