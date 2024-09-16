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

	err = e.cluster.CreateTopic("value_change", 3, 2)
	if err != nil {
		log.Fatalf("Failed to create the topic: %v", err)
	}

	err = e.cluster.CreateTopic("item_create", 3, 2)
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

func (e *Executer) RunConsumer() {
	_, err := e.cluster.CreateConsumer()
	if err != nil {
		log.Fatalf("Failed to create Kafka Consumer Group: %v\n", err)
	}
	defer e.wg.Done()
	defer func() {
		if err := e.cluster.Consumer.Close(); err != nil {
			log.Fatalf("Failed to close Kafka Consumer Group: %v", err)
		}
		color.Red("Kafka Consumer Group successfully closed")
	}()

	wgConsumer := new(sync.WaitGroup)
	wgConsumer.Add(1)

	//	run multiple consumers here and how they deal with the code

	go e.cluster.ListenForMessagesFromMulitpleTopics([]string{"order_details", "payment_confirmed"}, wgConsumer, e.ctx)

	<-e.ctx.Done()

	wgConsumer.Wait()
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

func main() {
	if len(os.Args) < 3 {
		color.Red("No Command Provided")
		return
	}
	command := os.Args[2]
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
		executer.SetupKafka()
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
