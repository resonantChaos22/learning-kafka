# Build and bring up the necessary services
up_build:
	@echo "Building and starting kafka..."
	docker-compose down -v
	docker-compose up --build -d
	@echo "Waiting for containers to start running normally..."
	sleep 10
	go run . kafka items setup
	@echo "Kafka Started"

# Run the Kafka producer (assuming you're running it from your Go service)
run_producer:
	@echo "Running producers..."
	go run . kafka items run-producer

# Run the Kafka consumer (assuming you're running it from your Go service)
run_consumer:
	@echo "Running consumer..."
	go run . kafka items run-consumer

# Run the Kafka consumer (assuming you're running it from your Go service)
stream:
	@echo "Running server to stream data..."
	go run . kafka items stream


# Tear down the environment
down:
	@echo "Shutting down services..."
	docker-compose down -v
