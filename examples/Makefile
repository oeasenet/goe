# GOE Examples Makefile

.PHONY: help start stop clean hello-world todo-api custom-module production-essentials code-first-config cdn-trusted-proxy

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Docker commands
start: ## Start Docker services (Redis, MongoDB, PostgreSQL)
	docker compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 3
	@echo "Services started!"

stop: ## Stop Docker services
	docker compose down

clean: ## Stop services and remove volumes
	docker compose down -v

logs: ## Show Docker logs
	docker compose logs -f

# Example runners
hello-world: ## Run the hello-world example
	cd 01-hello-world && go run .

todo-api: start ## Run the todo-api example (starts Docker first)
	cd 02-todo-api && go run .

custom-module: ## Run the custom-module example
	cd 04-custom-module && go run .

production-essentials: ## Run the production-essentials example
	cd 05-production-essentials && go run .

code-first-config: ## Run the code-first-config example
	cd 06-code-first-config && go run .

cdn-trusted-proxy: ## Run the cdn-trusted-proxy example (requires internet)
	cd 07-cdn-trusted-proxy && go run .

# Test commands
test-hello: ## Test hello-world endpoints
	@echo "Testing root endpoint..."
	curl -s http://localhost:3000/
	@echo "\n\nTesting hello endpoint..."
	curl -s http://localhost:3000/hello/GOE
	@echo "\n\nTesting health endpoint..."
	curl -s http://localhost:3000/health | jq .
	@echo "\n\nTesting echo endpoint..."
	curl -s -X POST http://localhost:3000/echo -H "Content-Type: application/json" -d '{"message":"hello"}' | jq .

test-todo: ## Test todo-api endpoints
	@echo "Creating a todo..."
	curl -s -X POST http://localhost:3000/todos -H "Content-Type: application/json" -d '{"title":"Learn GOE","description":"Study the framework"}' | jq .
	@echo "\n\nListing todos..."
	curl -s http://localhost:3000/todos | jq .
	@echo "\n\nGetting stats..."
	curl -s http://localhost:3000/todos/stats/summary | jq .

test-module: ## Test custom-module endpoints
	@echo "Testing health..."
	curl -s http://localhost:3000/health | jq .
	@echo "\n\nTesting metrics..."
	curl -s http://localhost:3000/metrics | jq .
	@echo "\n\nRecording demo request..."
	curl -s http://localhost:3000/demo | jq .
	@echo "\n\nChecking stats..."
	curl -s http://localhost:3000/stats | jq .
