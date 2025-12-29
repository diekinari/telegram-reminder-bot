.PHONY: build run dev docker-build docker-up docker-down clean

# Build the application
build:
	go build -o bot ./cmd/bot

# Run locally (requires .env file)
run: build
	./bot

# Development mode with hot reload (requires air)
dev:
	air

# Build docker image
docker-build:
	docker-compose build

# Start all services
docker-up:
	docker-compose up -d

# Stop all services
docker-down:
	docker-compose down

# View logs
docker-logs:
	docker-compose logs -f bot

# Clean build artifacts
clean:
	rm -f bot
	docker-compose down -v

# Run go mod tidy
tidy:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run tests
test:
	go test -v ./...
