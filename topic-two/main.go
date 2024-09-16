package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"topic-one/kafka"

	"github.com/IBM/sarama"
	"github.com/fatih/color"
)

const (
	DEBEZIUM_CONNECT_URL = "http://localhost:8083/connectors"
)

func SetupKafka(kc *kafka.KafkaCluster) {
	err := kc.CreateAdmin()

	if err != nil {
		log.Fatalf("Failed to create Kafka ClusterAdmin: %v\n", err)
	}

	defer func() {
		if err := kc.Admin.Close(); err != nil {
			log.Fatalf("Failed to close Kafka ClusterAdmin: %v", err)
		}
		color.Red("Kafka ClusterAdmin successfully closed!")
	}()

	err = kc.CreateTopic("order_details", 3, 2)
	if err != nil {
		log.Fatalf("Failed to create the topic: %v", err)
	}

	err = kc.CreateTopic("payment_confirmed", 3, 2)
	if err != nil {
		log.Fatalf("Failed to create the topic: %v", err)
	}

	func() {
		startTime := time.Now() // Capture the start time

		for {
			checkTime := time.Now()
			err := kc.ListTopics()
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

func RunProducer(kc *kafka.KafkaCluster, wg *sync.WaitGroup, ctx context.Context) {
	err := kc.CreateProducer()
	if err != nil {
		log.Fatalf("Failed to create Kafka Producer: %v\n", err)
	}
	defer wg.Done()
	defer func() {
		if err := kc.Producer.Close(); err != nil {
			log.Fatalf("Failed to close Kafka Producer: %v", err)
		}
		color.Red("Kafka Producer successfully closed")
	}()

	wgProducer := new(sync.WaitGroup)
	wgProducer.Add(2)

	go kc.SendDummyMessages("order_details", 5, wgProducer, ctx)
	go kc.SendDummyMessages("payment_confirmed", 2, wgProducer, ctx)

	<-ctx.Done()

	wgProducer.Wait()
}

func RunConsumer(kc *kafka.KafkaCluster, wg *sync.WaitGroup, ctx context.Context) {
	err := kc.CreateConsumer()
	if err != nil {
		log.Fatalf("Failed to create Kafka Consumer Group: %v\n", err)
	}
	defer wg.Done()
	defer func() {
		if err := kc.Consumer.Close(); err != nil {
			log.Fatalf("Failed to close Kafka Consumer Group: %v", err)
		}
		color.Red("Kafka Consumer Group successfully closed")
	}()

	wgConsumer := new(sync.WaitGroup)
	wgConsumer.Add(1)

	go kc.ListenForMessagesFromMulitpleTopics([]string{"order_details", "payment_confirmed"}, wgConsumer, ctx)

	<-ctx.Done()

	wgConsumer.Wait()
}

func SetupDebezium(kc *kafka.KafkaCluster) {
	isConnectorPresent, err := kc.CheckDebeziumConnector(DEBEZIUM_CONNECT_URL, "pg_connector")
	if err != nil {
		log.Fatalf("Error in checking presence of connector - %v", err)
	}

	if !isConnectorPresent {
		err = kc.CreateDebeziumConnector(DEBEZIUM_CONNECT_URL)
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
	defer cancel()

	switch command {
	case "setup":
		SetupKafka(kc)
	case "run-producer":
		wg.Add(1)
		go RunProducer(kc, wg, ctx)
	case "run-consumer":
		wg.Add(1)
		go RunConsumer(kc, wg, ctx)
	case "add-debezium":
		SetupDebezium(kc)
	default:
		log.Println("Command Not Found")
	}

	sig := <-sigChan
	color.Red("Received signal: %v, initiating graceful shutdown.", sig)

	cancel()

	wg.Wait()

	color.Red("Program Exited")
}
