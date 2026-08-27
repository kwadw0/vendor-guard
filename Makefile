.PHONY: graphify graphify-force build vet swag

graphify:
	@./scripts/graphify-update.sh

graphify-force:
	@./scripts/graphify-update.sh --force

build:
	go build -o preuvio ./cmd

vet:
	go vet ./...

swag:
	swag init -g cmd/main.go -o docs

sqlc:
	sqlc generate
