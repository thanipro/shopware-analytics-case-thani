.PHONY: help build start up down logs clean test test-go test-php install-deps

help:
	@echo "Shopware Analytics - Available Commands"
	@echo "========================================"
	@echo "make build        - Build all Docker containers"
	@echo "make start        - Build and start all services and show URLs"
	@echo "make up           - Start all services"
	@echo "make down         - Stop all services"
	@echo "make logs         - View logs from all services"
	@echo "make clean        - Remove all containers, volumes, and data"
	@echo "make test         - Run all tests"
	@echo "make test-go      - Run Go tests"
	@echo "make test-php     - Run PHP tests"
	@echo "make install-deps - Install dependencies locally"

build:
	docker-compose build

start:
	@mkdir -p data
	docker-compose build
	@docker-compose up -d
	@echo ""
	@echo "✓ Services started successfully!"
	@echo "================================"
	@echo ""
	@echo "Backend API:    http://localhost:8080"
	@echo "Analytics API:  http://localhost:8000"
	@echo "Frontend:       http://localhost:3000"
	@echo ""
	@echo "View logs:      make logs"
	@echo "Stop services:  make down"
	@echo ""

up:
	@mkdir -p data
	docker-compose up -d
	@echo ""
	@echo "Services started successfully!"
	@echo "================================"
	@echo "Ingestion API:  http://localhost:8080"
	@echo "Analytics API:  http://localhost:8000"
	@echo "Frontend:       http://localhost:3000"
	@echo ""
	@echo "View logs with: make logs"

down:
	docker-compose down

logs:
	docker-compose logs -f

clean:
	docker-compose down -v
	rm -rf data/*.db
	@echo "All containers, volumes, and data cleaned"

test: test-go test-php
	@echo "All tests completed"

test-go:
	@echo "Running Go backend tests..."
	cd backend && go test -v ./...

test-php:
	@echo "Running PHP analytics tests..."
	cd analytics && composer install && vendor/bin/phpunit

install-deps:
	@echo "Installing Go dependencies..."
	cd backend && go mod download
	@echo "Installing PHP dependencies..."
	cd analytics && composer install
	@echo "Installing Node dependencies..."
	cd frontend && npm install
	@echo "All dependencies installed"
