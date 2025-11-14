BINARY=threat-feed
DOCKER_IMAGE=threat-feed
DOCKER_TAG=latest
CONTAINER_NAME=threat-feed

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

build-docker:
	@echo "--> Building Docker image"
	docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .

run-docker:
	@echo "--> Running Docker container (production mode)"
	docker run -d \
		--name ${CONTAINER_NAME} \
		-p 8080:8080 \
		-e GIN_MODE=release \
		-e RATE_LIMIT=1000 \
		-e RATE_BURST=10 \
		-e ENABLE_PPROF=false \
		${DOCKER_IMAGE}:${DOCKER_TAG}
	@echo "Container started: ${CONTAINER_NAME}"
	@echo "API available at: http://localhost:8080"


run-docker-dev:
	@echo "--> Running Docker container (development mode with pprof)"
	docker run -d \
		--name ${CONTAINER_NAME}-dev \
		-p 8080:8080 \
		-p 6060:6060 \
		-e GIN_MODE=debug \
		-e RATE_LIMIT=1000 \
		-e RATE_BURST=10 \
		-e ENABLE_PPROF=true \
		${DOCKER_IMAGE}:${DOCKER_TAG}
	@echo "Container started: ${CONTAINER_NAME}-dev"
	@echo "API available at: http://localhost:8080"
	@echo "pprof available at: http://localhost:6060/debug/pprof/"
