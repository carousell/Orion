ifdef RACE
	OPTS = -race -v
	else
	OPTS = -v
endif

all: clean vet test build

cleanall: clean dockerclean

ci: clean vet bench build

vet:
	go vet ./ServiceName/...

lint:
	golint ./ServiceName/...

test:
	go test -cover ./ServiceName/...

build:
	go build $(OPTS) ./ServiceName/...

build-linux:
	GOOS=linux GOARCH=amd64 go build $(OPTS) ./ServiceName/...

bench:
	go test -cover --bench ./ServiceName/... ./ServiceName/...

benchmark: bench

clean:
	go clean ./...

race:
	RACE=true make all

doc:
	godoc -http=:6060

update:
	govendor fetch +vendor
	go get -u github.com/carousell/Orion/protoc-gen-orion/v2

run:
	exec ./run.sh

proto:
	cd ServiceName/ServiceName_proto/; bash generate.sh

runclient:
	cd ServiceName/cmd/client; go build .; ./client

dockerclean:
	echo "remove exited containers"
	docker ps --filter status=dead --filter status=exited -aq | xargs  docker rm -v
	docker images --no-trunc | grep "<none>" | awk '{print $3}' | xargs  docker rmi
	echo "^ above errors are ok"

install: macinstall goinstall

goinstall:
	go get -u github.com/kardianos/govendor
	go get -u github.com/shurcooL/Go-Package-Store/cmd/Go-Package-Store
	go get -u github.com/golang/lint/golint
	go get -u google.golang.org/grpc
	go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
	go get -u github.com/carousell/Orion/protoc-gen-orion/v2

macinstall:
	brew install protobuf
