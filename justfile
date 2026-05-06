default:
    just --list

run:
    go run cmd/bot/main.go

build:
    go build -o bin/bot cmd/bot/main.go

test:
    go test -v -race -cover ./...

lint:
    golangci-lint run --fix

fmt:
    go fmt ./...
    goimports -l -w .

check:
    just fmt
    just lint
    just test
    just build
