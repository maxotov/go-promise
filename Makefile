all: build

test:
	go test -race -p 1 -bench=. -cover -v ./...

build:
	CGO_ENABLED=0 go build -o go-promise
