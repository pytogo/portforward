all: clean build
.PHONY: all clean build

clean:
	rm portfoward.so portfoward.h

build:
	go build -buildmode=c-shared -o portforward.so portforward.go
