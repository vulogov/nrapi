all: compile
pre:
	go get github.com/Jeffail/gabs/v2
c:
	go build  -v ./...
test:
	go test -v
rebuild: pre c test
compile: c test
