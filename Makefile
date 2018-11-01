GOCMD=go
GOBUILD=$(GOCMD) build
GOFMT=$(GOCMD) fmt
GOCLEAN=$(GOCMD) clean
BINARY_NAME=prj2
PORT=44461
HOSTFILE=hostfile

all: clean build
run: build start

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

start:
	$(BINARY_NAME) -h $(HOSTFILE) -p $(PORT)

leader-test2:
	$(BINARY_NAME) -h $(HOSTFILE) -p $(PORT) -t2

leader-test4:
	$(BINARY_NAME) -h $(HOSTFILE) -p $(PORT) -t4