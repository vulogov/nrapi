all: compile
pre:
	go get github.com/Jeffail/gabs/v2
	go get github.com/google/uuid
c:
	go build  -v ./...
test:
	go test -v
rebuild: pre c test
compile: c test
