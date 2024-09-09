package main

import (
	"log"
	"time"
	"topic-one/kafka"

	"github.com/IBM/sarama"
)

func main() {
	config := sarama.NewConfig()

	brokers := []string{"localhost:9092", "localhost:9093", "locahost:9094"}
	kc, err := kafka.NewKafkaCluster(brokers, config, sarama.DefaultVersion)
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

	// err = kc.DeleteTopic("order_details")
	// if err != nil {
	// 	log.Fatalf("Failed to delete the topic: %v", err)
	// }

}
