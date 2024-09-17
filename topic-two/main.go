package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"topic-two/items"
	"topic-two/kafka"

	"github.com/IBM/sarama"
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
}

func NewExecuter(cluster *kafka.KafkaCluster, wg *sync.WaitGroup, ctx context.Context, store items.Storage) *Executer {
	return &Executer{
		cluster: cluster,
		wg:      wg,
		ctx:     ctx,
		store:   store,
	}
}

func (e *Executer) SetupKafka() {
	err := e.cluster.CreateAdmin()

	if err != nil {
		log.Fatalf("Failed to create Kafka ClusterAdmin: %v\n", err)
	}

	defer func() {
		if err := e.cluster.Admin.Close(); err != nil {
			log.Fatalf("Failed to close Kafka ClusterAdmin: %v", err)
		}
		color.Red("Kafka ClusterAdmin successfully closed!")
	}()

	err = e.cluster.CreateTopic(VALUE_CHANGE_TOPIC, 3, 2)
	if err != nil {
		log.Fatalf("Failed to create the topic: %v", err)
	}

	func() {
		startTime := time.Now() // Capture the start time

		for {
			checkTime := time.Now()
			err := e.cluster.ListTopics()
			color.Magenta("Took %dms to list topics", time.Since(checkTime).Milliseconds())
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
		log.Fatalf("Error in checking presence of connector - %v", err)
	}

	if !isConnectorPresent {
		err = e.cluster.CreateDebeziumConnector(DEBEZIUM_CONNECT_URL)
		if err != nil {
			log.Fatalf("Error in creating the connector - %v", err)
		}
	}

	log.Println("Debezium Connector Is Active!")
}

func (e *Executer) SetupDB() {
	err := e.store.CreateItemTable()
	if err != nil {
		log.Fatalf("Error in creating item table: %v", err)
	}

	items := []*items.Item{{Name: "Spotify", Value: 500.0}, {Name: "Reddit", Value: 800.0}, {Name: "Netflix", Value: 200.0}}

	for _, item := range items {
		err = e.store.CreateItem(item)
		if err != nil {
			log.Fatalf("Error in adding item to table: %v", err)
		}
	}
}

func (e *Executer) Setup() {
	e.SetupKafka()
	e.SetupDebezium()
	e.SetupDB()
}

func main() {
	if len(os.Args) < 4 {
		color.Red("No Command Provided")
		return
	}
	command := os.Args[3]
	color.Yellow(command)

	brokers := []string{"localhost:9092", "localhost:9093", "locahost:9094"}
	kc := kafka.NewKafkaCluster(brokers, sarama.DefaultVersion)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	wg := new(sync.WaitGroup)

	ctx, cancel := context.WithCancel(context.Background())
	store, err := items.NewPostgresStore()
	if err != nil {
		log.Fatalf("Error in creating Postgres Store: %v", err)
	}
	executer := NewExecuter(kc, wg, ctx, store)
	defer cancel()

	switch command {
	case "setup":
		executer.Setup()
	case "run-producer":
		executer.wg.Add(1)
		go executer.RunProducer()
	case "run-consumer":
		executer.wg.Add(1)
		go executer.RunConsumer()
	case "add-debezium":
		executer.SetupDebezium()
	default:
		log.Println("Command Not Found")
	}

	sig := <-sigChan
	color.Red("Received signal: %v, initiating graceful shutdown.", sig)

	cancel()

	executer.wg.Wait()

	color.Red("Program Exited")
}
