.PHONY: build test run docker-build docker-run clean

# Build the application
build:
	go build -o bin/main cmd/auction/main.go

# Run tests
test:
	go test ./...

# Run the application locally
run:
	go run cmd/auction/main.go

# Build Docker image
docker-build:
	docker build -t auction-app .

# Run Docker container
docker-run:
	docker run -d -p 3000:3000 --network auction-network --name auction-server --env-file .env auction-server

# Start development environment with Docker Compose
dev:
	docker-compose up --build

# Run database migrations
migrate:
	go run cmd/migration/main.go

# Clean up build artifacts
clean:
	rm -rf bin/

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	go fmt ./...
