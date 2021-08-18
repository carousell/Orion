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
	go test -cover -race -coverprofile=coverage.txt -covermode=atomic --bench ./orion/... ./utils/... ./interceptors/...

benchmark: bench

clean:
	go clean ./...

race:
	RACE=true make all

doc:
	godoc -http=:6060

list-updates:
	go list -u -m all

update:
	go get -u github.com/carousell/Orion/protoc-gen-orion
	go get -u -m
	go mod tidy
	go mod vendor

run:
	exec ./run.sh

dockerclean:
	echo "remove exited containers"
	docker ps --filter status=dead --filter status=exited -aq | xargs  docker rm -v
	docker images --no-trunc | grep "<none>" | awk '{print $3}' | xargs  docker rmi
	echo "^ above errors are ok"

install: macinstall goinstall

goinstall:
	go get -u github.com/shurcooL/Go-Package-Store/cmd/Go-Package-Store
	go get -u github.com/golang/lint/golint
	go get -u google.golang.org/grpc
	go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
	go get -u github.com/carousell/Orion/protoc-gen-orion

macinstall:
	brew install protobuf

gen:
	go generate ./orion ./utils ./interceptors ./example

sonar-test:
	$(GOVENDORCMD) test $(OPTS) ./orion/... -coverprofile coverage.out -coverpkg=./orion/...

sonar-check-pr:
	sonar-scanner \
		-Dproject.settings=sonar.properties \
		-Dsonar.host.url=http://qa-sonarqube.carouselltech.com \
		-Dsonar.login=${SONAR_LOGIN_TOKEN} \
		-Dsonar.pullrequest.key=${TRAVIS_PULL_REQUEST} \
		-Dsonar.pullrequest.branch=${TRAVIS_PULL_REQUEST_BRANCH} \
		-Dsonar.pullrequest.base=master

sonar-check-baseline:
	sonar-scanner \
		-Dproject.settings=sonar.properties \
		-Dsonar.host.url=http://qa-sonarqube.carouselltech.com \
		-Dsonar.login=${SONAR_LOGIN_TOKEN}
