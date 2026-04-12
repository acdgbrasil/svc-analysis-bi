.PHONY: build test lint dev clean docker-build vet

APP_NAME := svc-analysis-bi
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server/

test:
	go test -race -cover -count=1 ./...

vet:
	go vet ./...

lint: vet

dev:
	go run ./cmd/server/

clean:
	rm -rf $(BUILD_DIR)

docker-build:
	docker build -t $(APP_NAME):local .
