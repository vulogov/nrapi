all: compile
pre:
	go get github.com/Jeffail/gabs/v2
	go get github.com/google/uuid
	go get -u github.com/rocketlaunchr/dataframe-go/...
c:
	go build  -v ./...
test:
	go test -v
rebuild: pre c test
compile: c test
