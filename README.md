# Orion [![Build Status](https://travis-ci.com/carousell/Orion.svg?token=kSVweyyqayUyyfutjTqD&branch=master)](https://travis-ci.com/carousell/Orion) [![Go Report Card](https://goreportcard.com/badge/github.com/carousell/Orion)](https://goreportcard.com/report/github.com/carousell/Orion) [![GoDoc](https://godoc.org/github.com/carousell/Orion/orion?status.svg)](https://godoc.org/github.com/carousell/Orion/orion)

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
go get -u github.com/golang/lint/golint
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

## Project Status
Orion is in use at production at Carousell and powers multiple (40+) services serving thousands of requests per second,
we ensure all updates are backward compatible unless it involves a major bug or security issue.

## License
This code is available under the following https://github.com/carousell/Orion/blob/master/LICENSE
