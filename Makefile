include .env

TARGET = bingsoo
SOURCE = cmd/bingsoo/main.go
DEPENDENCIES = postgres postgres.init redis

.PHONY: all build run dependencies test format

all: format run

build:
	@echo "==> Compiling code.."
	go build -o ${TARGET} ${SOURCE}

run:
	@echo "==> Executing code.."
	@go run ${SOURCE} \
		--port 8080 \
		--slack-access-token ${SLACK_ACCESS_TOKEN} \
		--slack-signing-secret ${SLACK_SIGNING_SECRET}

dependencies:
	@echo "==> Starting auxiliary containers.."
	docker-compose up -d ${DEPENDENCIES}

test:
	@echo "==> Running tests.."
	go test -v ./...

format:
	@echo "==> Formatting code.."
	gofmt -w .
