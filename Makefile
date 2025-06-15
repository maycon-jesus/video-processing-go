up:
	docker compose up
down:
	docker compose down
dev:
	docker exec -it video-processing-go-app-1 bash

# Test targets
test:
	go test ./internal

test-bench:
	go test -bench=. ./internal

test-cover:
	go test -cover ./internal

test-cover-html:
	go test -coverprofile=coverage.out ./internal
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-verbose:
	go test -v ./internal

test-all: test test-bench test-cover