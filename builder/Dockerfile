FROM golang:1.16

RUN go get github.com/derekparker/delve/cmd/dlv

RUN mkdir -p /go/src/github.com/carousell/ServiceName
RUN mkdir -p /opt/config/

COPY . /go/src/github.com/carousell/ServiceName
COPY ./ServiceName/ServiceName.toml /opt/config/

RUN go install github.com/carousell/ServiceName/ServiceName/cmd/server

EXPOSE 9281 9282 9283 9284
