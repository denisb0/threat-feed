BINARY=njord

all: clean build

clean:
	@echo "--> Target directory clean up"
	rm -rf ./.build/target
	rm -f ${BINARY}

build:
	go build -o ${BINARY} ./cmd

test-unit:
	@echo "--> Running unit tests"
	go test -v -race -cover ./...

lint:
	@echo "--> Running linters"
	go vet ./...
	go fmt ./...