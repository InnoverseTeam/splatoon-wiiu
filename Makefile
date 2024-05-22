BINARY_NAME=serveur-splatoon
MAIN_FILE=init.go

.PHONY: all build run clean

all: build

build:
	go build -o $(BINARY_NAME) $(MAIN_FILE)

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
