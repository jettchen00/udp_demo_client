BINARY_NAME=udpClient

all: build

build:
	go build -o ${BINARY_NAME} main.go

clean:
	rm ${BINARY_NAME}