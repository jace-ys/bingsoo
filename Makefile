include .env

TARGET = bingsoo
SOURCE = cmd/bingsoo/main.go
DEPENDENCIES = postgres postgres.init redis

.PHONY: all build run proxy dependencies test format

all: format run

build:
	@echo "==> Compiling code.."
	go build -o ${TARGET} ${SOURCE}

run:
	@echo "==> Executing code.."
	@go run ${SOURCE} \
		--port 8080 \
		--concurrency 4 \
		--slack-access-token ${SLACK_ACCESS_TOKEN} \
		--slack-signing-secret ${SLACK_SIGNING_SECRET}

proxy:
	ngrok http 8080

dependencies:
	@echo "==> Starting auxiliary containers.."
	docker-compose up -d ${DEPENDENCIES}

test:
	@echo "==> Running tests.."
	go test -v ./...

format:
	@echo "==> Formatting code.."
	gofmt -w .
