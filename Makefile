run:
	go run cmd/server/main.go

swag:
	swag init -g cmd/server/main.go -o ./api/docs

lint:
	golangci-lint run

dev:
	air

build:
	go build -o bin/server cmd/server/main.go