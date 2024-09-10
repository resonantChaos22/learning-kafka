package main

import (
	"log"
	"os"
	"time"
	"topic-one/kafka"

	"github.com/IBM/sarama"
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
		log.Println("Kafka ClusterAdmin successfully closed!")
	}()

	err = kc.CreateTopic("order_details", 3, 2)
	if err != nil {
		log.Fatalf("Failed to create the topic: %v", err)
	}

	func() {
		startTime := time.Now() // Capture the start time

		for {
			checkTime := time.Now()
			err := kc.ListTopics()
			log.Printf("Took %dms to list topics", time.Since(checkTime).Milliseconds())
			if err == nil {
				elapsed := time.Since(startTime) // Calculate elapsed time
				log.Printf("It took %d milliseconds to achieve sync", elapsed.Milliseconds())
				return
			}
			time.Sleep(2 * time.Millisecond)
		}
	}()
}

func RunProducer(kc *kafka.KafkaCluster) {
	err := kc.CreateProducer()
	if err != nil {
		log.Fatalf("Failed to create Kafka Producer: %v\n", err)
	}

	kc.SendDummyMessages()

}

func main() {

	if len(os.Args) < 3 {
		log.Println("No Command Provided")
		return
	}
	command := os.Args[2]
	log.Println(command)

	brokers := []string{"localhost:9092", "localhost:9093", "locahost:9094"}
	kc := kafka.NewKafkaCluster(brokers, sarama.DefaultVersion)

	switch command {
	case "setup":
		SetupKafka(kc)
	case "run-producer":
		RunProducer(kc)
	default:
		log.Println("Command Not Found")
	}

}
