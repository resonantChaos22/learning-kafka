package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"topic-two/cmd"
	"topic-two/kafka"

	"github.com/IBM/sarama"
	"github.com/fatih/color"
)

func main() {
	if len(os.Args) < 2 {
		color.Red("No Command Provided")
		return
	}
	command := os.Args[1]
	color.Yellow(command)

	brokers := []string{"localhost:9092", "localhost:9093", "locahost:9094"}
	kc := kafka.NewKafkaCluster(brokers, sarama.DefaultVersion)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)

	ctx, cancel := context.WithCancel(context.Background())
	executer := cmd.NewExecuter(kc, nil, ctx, nil)
	defer cancel()

	switch command {
	case "setup":
		executer.Setup()
		color.Red("Program exited!")
		return
	case "run-producer":
		wg := new(sync.WaitGroup)
		executer.SetWg(wg)
		go executer.RunProducer()
	case "run-consumer":
		wg := new(sync.WaitGroup)
		executer.SetWg(wg)
		go executer.RunConsumer()
	case "add-debezium":
		executer.SetupDebezium()
		color.Red("Program exited!")
		return
	case "stream":
		wg := new(sync.WaitGroup)
		executer.SetWg(wg)
		go executer.Stream()
	default:
		log.Println("Command Not Found")
	}

	select {
	case sig := <-sigChan:
		color.Red("Received signal: %v, initiating graceful shutdown.", sig)
	case err := <-executer.GetErrors():
		if err != nil {
			color.Red("Received error: %s, initiating graceful shutdown.", err.Error())
		} else {
			color.Yellow("Error channel closed, initiating graceful shutdown.")
		}
	}

	cancel()
	executer.CloseErrorChannel()

	executer.Wait()

	color.Red("Program Exited")
}
