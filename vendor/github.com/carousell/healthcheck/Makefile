ifdef RACE
        OPTS = -race -v
        else
        OPTS = -v
endif

all: clean vet test build

cleanall: clean dockerclean

ci: clean vet bench build

vet:
	go vet .

lint:
	golint .

test:
	go test -cover --bench . .

build:
	go build $(OPTS) ./...

build-linux:
	GOOS=linux GOARCH=amd64 go build $(OPTS) ./...

bench:
	go test -cover --bench .  .

benchmark: bench

clean:
	go clean ./...

race:
	RACE=true make all

doc:
	godoc -http=:6060

dockerclean:
	echo "remove exited containers"
	docker ps --filter status=dead --filter status=exited -aq | xargs  docker rm -v
	docker images --no-trunc | grep "<none>" | awk '{print $3}' | xargs  docker rmi
	echo "^ above errors are ok"
