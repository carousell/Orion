# Orion [![Build Status](https://travis-ci.com/carousell/Orion.svg?token=kSVweyyqayUyyfutjTqD&branch=master)](https://travis-ci.com/carousell/Orion) [![Go Report Card](https://goreportcard.com/badge/github.com/carousell/Orion)](https://goreportcard.com/report/github.com/carousell/Orion) [![codecov](https://codecov.io/gh/carousell/Orion/branch/master/graph/badge.svg?token=XEOedAF3IG)](https://codecov.io/gh/carousell/Orion) [![GoDoc](https://godoc.org/github.com/carousell/Orion/orion?status.svg)](https://godoc.org/github.com/carousell/Orion/orion)

Orion is a small lightweight framework written around grpc/protobuf with the aim to shorten time to build microservices at Carousell.

It is derived from 'Framework' a small microservices framework written and used inside https://carousell.com, It comes with a number of sensible defaults such as zipkin tracing, hystrix, live reload of configuration, etc.

## Getting Started
Follow the guide at https://github.com/carousell/Orion/blob/master/orion/README.md

## Setup Instructions
Orion is written in golang, please follow instructions on [https://golang.org/doc/install](https://golang.org/doc/install) to install, or you can also run
```
brew install golang
```
or
```
sudo dnf install golang
```
Make sure you are on go 1.9 or later
add the following lines to your `~/.profile`
```
export GOPATH="$HOME/code/go"
export GOBIN="$GOPATH/bin"
export PATH="$GOBIN:$PATH"
export PATH="$HOME/.gotools:$PATH"
```

source your `~/.profile`
```
source ~/.profile
```

then create the code dir
```
mkdir -p $GOPATH
```

we use `govendor` to vendor package in Orion, install it by running
```
go get -u github.com/kardianos/govendor
```
another helpful tool to check for unupdated packages is `Go-Package-Store`, install it by running
```
go get -u github.com/shurcooL/Go-Package-Store/cmd/Go-Package-Store
```
now clone this repo
```
mkdir -p $GOPATH/src/github.com/carousell/
git clone git@github.com:carousell/Orion.git $GOPATH/src/github.com/carousell/Orion
```

You need the following tools to better develop for go
```
go get -u golang.org/x/lint/golint
```

now you can build the package by using `make build`

## gRPC
for gRPC, you need to follow the following steps

get gRPC codebase
```
go get -u google.golang.org/grpc
```

install protobuf
```
brew install protobuf
```

install the protoc plugin for go
```
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
```

install the protoc plugin for orion
```
go get -u github.com/carousell/Orion/protoc-gen-orion
```

## protoc-gen-orion

### Installation

Install the binary from source for golang version >= 1.17.
```bash
go install github.com/carousell/Orion/protoc-gen-orion@latest
```
Or, golang version < 1.17.
```bash
go get github.com/carousell/Orion/protoc-gen-orion 
```

Install the dependencies tools.
- `protoc-gen-go` is for generating the message structure which has two different repo.
  - For version lower and equal than v1.5.2
    ```bash
    go install github.com/golang/protobuf@latest
    ```
  - For version higher and equal than v1.20.0.
    ```bash
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    ```
- `protoc-gen-go-grpc` is for generating the grpc service which works with the version of `protoc-gen-go` higher and equal than v1.20.0
  ```bash
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```

### Usage

#### Working with `protocgen` generation tool (shared-proto).
Protogen generation tool is used for generating protobuffer for services using `shared-proto` repository. 
It helps in generating protobuffer within the service itself.
Please refer to this [link](https://carousell.atlassian.net/wiki/spaces/RFC/pages/434471178/How+to+onboard+your+service+to+protogen) for more details.

**This tool is only used for services using `shared-proto` repository.**

#### Working with `<bu>-proto-gen-go` generation tool (service-proto).
`<bu>-proto-gen-go` generation tool is used for generating protobuffer for services using `<bu>-proto` repository.
This tool automatically generates the protobuffer for the service and also generates the go mod to be imported for the
generated protobuffers.
Please refer to this [link](https://carousell.atlassian.net/wiki/spaces/CTF/pages/2216689780/Protobuffer+Management+Guide) for more details.

## Project Status
Orion is in use at production at Carousell and powers multiple (100+) services serving thousands of requests per second,
we ensure all updates are backward compatible unless it involves a major bug or security issue.

## License
This code is available under the following https://github.com/carousell/Orion/blob/master/LICENSE
