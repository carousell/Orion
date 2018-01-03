ifdef RACE
	OPTS = -race -v
	else
	OPTS = -v
endif

all: clean vet test build

cleanall: clean dockerclean

ci: clean vet bench build

vet:
	go vet ./orion/... ./utils/...

lint:
	golint ./orion/... ./utils/...

test:
	go test -cover ./orion/... ./utils/...

build:
	go build $(OPTS) ./orion/... ./utils/...

build-linux:
	GOOS=linux GOARCH=amd64 go build $(OPTS) ./orion/...

bench:
	go test -cover --bench ./orion/... ./utils/...

benchmark: bench

clean:
	go clean ./...

race:
	RACE=true make all

doc:
	godoc -http=:6060

update:
	govendor fetch +vendor
	go get -u github.com/carousell/Orion/protoc-gen-orion

run:
	exec ./run.sh

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
	go get -u github.com/carousell/Orion/protoc-gen-orion

macinstall:
	brew install protobuf

gen:
	go generate ./orion ./utils ./interceptors
