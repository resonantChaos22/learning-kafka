# How To Run

Please do make sure that `Docker` is installed and is running.

## Starting Kafka

- Run `make up_build` in terminal to start the Kafka Cluster and add the required topics.

## Run Consumer

- Run `make run_consumer` in terminal to start listening for messages on the topics.

## Run Producer

- Run `make run_producer` in terminal to run the goroutines responsible for producing messages to the topics.

## Shutting Down Kafka

- Run `make down`
