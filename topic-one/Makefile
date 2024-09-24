# Build and bring up the necessary services
up_build:
	@echo "Building and starting kafka..."
	docker-compose down -v
	docker-compose up --build -d
	go run . items setup
	@echo "Kafka Started"

# Run the Kafka producer (assuming you're running it from your Go service)
run_producer:
	@echo "Running producer..."
	go run . items run-producer

# Run the Kafka consumer (assuming you're running it from your Go service)
run_consumer:
	@echo "Running consumer..."
	go run . items run-consumer

# Tear down the environment
down:
	@echo "Shutting down services..."
	docker-compose down -v
