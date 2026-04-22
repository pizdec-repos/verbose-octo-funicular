run:
    go run cmd/bot/main.go

build:
    go build -o bin/bot cmd/bot/main.go

test:
    go test -v -race -cover ./...

lint:
    golangci-lint run


