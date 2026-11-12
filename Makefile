SIZE ?= small

.PHONY: build run seed frontend test bench docker-up docker-seed docker-down

build:
	cd backend && go build -o ../bin/ ./cmd/...
	cd frontend && npm install && npm run build

run: backend/.env
	cd backend && ../bin/server

seed: backend/.env
	cd backend && go run ./cmd/seed -size $(SIZE)

frontend:
	cd frontend && npm install && npm run dev

test:
	cd backend && go test ./...

docker-up: backend/.env
	docker compose up -d --build

docker-seed:
	docker compose exec backend seed -size $(SIZE)

docker-down:
	docker compose down

backend/.env:
	cp backend/.env.example backend/.env
