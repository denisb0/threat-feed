BINARY=threat-feed

all: clean build

clean:
	rm -f ${BINARY}

build:
	go build -o ${BINARY} 

test-unit:
	@echo "--> Running unit tests"
	go test -v -race -cover ./...

lint:
	@echo "--> Running linters"
	go vet ./...
	go fmt ./...