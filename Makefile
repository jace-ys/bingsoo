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
		--signing-secret ${SIGNING_SECRET} \
		--database-url postgres://bingsoo:bingsoo@127.0.0.1:5432/bingsoo?sslmode=disable \
		--redis-url redis://127.0.0.1:6379

image:
	@echo "==> Building image.."
	docker-compose build bingsoo

proxy:
	@echo "==> Starting proxy.."
	@ngrok http 8080

dependencies:
	@echo "==> Starting auxiliary containers.."
	docker-compose up -d ${DEPENDENCIES}

test:
	@echo "==> Running tests.."
	go test -v ./...

format:
	@echo "==> Formatting code.."
	go fmt ./...
