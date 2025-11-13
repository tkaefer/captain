BINARY=captain

all: build

build:
	go build -o $(BINARY) ./cmd/captain

clean:
	rm -f $(BINARY)
