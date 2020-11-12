module github.com/carousell/Orion

go 1.15

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0

require (
	cloud.google.com/go v0.71.0 // indirect
	cloud.google.com/go/pubsub v1.3.1
	github.com/DataDog/datadog-go v0.0.0-20180822151419-281ae9f2d895 // indirect
	github.com/DataDog/zstd v0.0.0-20160706220725-2bf71ec48360 // indirect
	github.com/RichardKnop/logging v0.0.0-20181101035820-b1d5d44c82d6 // indirect
	github.com/RichardKnop/machinery v0.0.0-20190125102247-b25a799bf62a
	github.com/Shopify/sarama v0.0.0-20190123165648-e775ee1118ac
	github.com/Shopify/toxiproxy v2.1.4+incompatible // indirect
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/apache/thrift v0.12.0 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/bugsnag/bugsnag-go v1.4.0
	github.com/bugsnag/panicwrap v0.0.0-20180510051541-1d162ee1264c // indirect
	github.com/cactus/go-statsd-client/statsd v0.0.0-20190805010426-5089fcbbe532 // indirect
	github.com/carousell/log-svc-client v0.0.0-20201105074630-f7c9eb3ae577
	github.com/certifi/gocertifi v0.0.0-20190105021004-abcd57078448 // indirect
	github.com/eapache/go-resiliency v0.0.0-20181214162408-487be0453c7b // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v0.0.0-20180227141424-093482f3f8ce // indirect
	github.com/elastic/go-sysinfo v1.1.0 // indirect
	github.com/fortytw2/leaktest v1.3.0
	github.com/getsentry/raven-go v0.0.0-20190125112653-238ebd86338d
	github.com/go-kit/kit v0.8.0
	github.com/go-logfmt/logfmt v0.4.0 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/golang-migrate/migrate v0.0.0-20180905021119-16f2b1736e65
	github.com/golang/protobuf v1.4.3
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/gorilla/mux v1.7.0
	github.com/gorilla/websocket v0.0.0-20181206070239-95ba29eb981b
	github.com/grpc-ecosystem/go-grpc-middleware v0.0.0-20190118093823-f849b5445de4
	github.com/grpc-ecosystem/go-grpc-prometheus v0.0.0-20181025070259-68e3a13e4117
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1 // indirect
	github.com/micro/protobuf v0.0.0-20180321161605-ebd3be6d4fdb
	github.com/mitchellh/mapstructure v1.1.2
	github.com/newrelic/go-agent v2.3.0+incompatible
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.0.2
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.3.5 // indirect
	github.com/openzipkin/zipkin-go-opentracing v0.3.5
	github.com/patrickmn/go-cache v0.0.0-20180815053127-5633e0862627
	github.com/pborman/uuid v1.2.0
	github.com/pierrec/lz4 v0.0.0-20190122092403-561038314e13 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.2
	github.com/prometheus/common v0.2.0 // indirect
	github.com/prometheus/procfs v0.0.3 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	github.com/santhosh-tekuri/jsonschema v1.2.4 // indirect
	github.com/satori/go.uuid v0.0.0-20181028125025-b2ce2384e17b
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/afero v1.2.1 // indirect
	github.com/spf13/cobra v0.0.4
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.3.2
	github.com/streadway/amqp v0.0.0-20181205114330-a314942b2fd9 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/stvp/rollbar v0.0.0-20171113052335-4a50daf855af
	go.elastic.co/apm v1.4.0
	golang.org/x/net v0.0.0-20201031054903-ff519b6c9102
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/sys v0.0.0-20201101102859-da207088b7d1 // indirect
	golang.org/x/text v0.3.4 // indirect
	google.golang.org/api v0.34.0
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20201104152603-2e45c02ce95c // indirect
	google.golang.org/grpc v1.33.1
	gopkg.in/airbrake/gobrake.v2 v2.0.9
)
