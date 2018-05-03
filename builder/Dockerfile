From golang:1.10

RUN go get github.com/derekparker/delve/cmd/dlv

RUN mkdir -p /go/src/github.com/carousell/Orion/builder
RUN mkdir -p /opt/config/

COPY . /go/src/github.com/carousell/Orion/builder
COPY ./ServiceName/ServiceName.toml /opt/config/

RUN go install github.com/carousell/Orion/builder/ServiceName/cmd/server

EXPOSE 9281 9282 9283 9284
