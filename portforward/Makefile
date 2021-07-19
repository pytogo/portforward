all: clean build
.PHONY: all clean build

clean:
	rm portforward.so portforward.h

build:
	go build -buildmode=c-shared -o portforward.so portforward.go
