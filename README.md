# Orion [![Build Status](https://travis-ci.com/carousell/Orion.svg?token=kSVweyyqayUyyfutjTqD&branch=master)](https://travis-ci.com/carousell/Orion)

## Setup Instructions
Orion is written in golang, please follow instructions on [https://golang.org/doc/install](https://golang.org/doc/install) to install, or you can also run
```
brew install golang
```
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

## Creating Service
just run ```./create.sh <service-name>```

## Using Orion
### Adding new API call
* Update Proto defination
* Add Service call in service/types.go
* Implement the service call
* Add Endpoints
* Update HTTP/gRPC transports
