GOCMD=go
GOBUILD=$(GOCMD) build
GOFMT=$(GOCMD) fmt
GOCLEAN=$(GOCMD) clean
BINARY_NAME=prj2

all: clean build

build: clean fmt
	$(GOBUILD) -o $(BINARY_NAME) -v

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run-docker: stop-docker clean build
	docker-compose up

stop-docker:
	docker-compose down

fmt:
	$(GOFMT)
