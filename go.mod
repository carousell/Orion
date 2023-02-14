module github.com/carousell/Orion

go 1.17

replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
	github.com/openzipkin-contrib/zipkin-go-opentracing => github.com/openzipkin-contrib/zipkin-go-opentracing v0.3.5
)

require (
	cloud.google.com/go v0.34.0
	github.com/DataDog/datadog-go v0.0.0-20180822151419-281ae9f2d895 // indirect
	github.com/RichardKnop/logging v0.0.0-20181101035820-b1d5d44c82d6 // indirect
	github.com/RichardKnop/machinery v0.0.0-20190125102247-b25a799bf62a
	github.com/Shopify/sarama v1.19.0
	github.com/Shopify/toxiproxy v2.1.4+incompatible // indirect
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/apache/thrift v0.0.0-20190131011427-2ec93c8a2da2 // indirect
	github.com/bitly/go-simplejson v0.5.0 // indirect
	github.com/bugsnag/bugsnag-go v1.4.0
	github.com/bugsnag/panicwrap v0.0.0-20180510051541-1d162ee1264c // indirect
	github.com/cactus/go-statsd-client/statsd v0.0.0-20190805010426-5089fcbbe532 // indirect
	github.com/certifi/gocertifi v0.0.0-20190105021004-abcd57078448 // indirect
	github.com/eapache/go-resiliency v1.2.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20180814174437-776d5712da21 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/elastic/go-sysinfo v1.1.0 // indirect
	github.com/fortytw2/leaktest v1.3.0
	github.com/frankban/quicktest v1.13.0 // indirect
	github.com/getsentry/raven-go v0.0.0-20190125112653-238ebd86338d
	github.com/go-kit/kit v0.8.0
	github.com/go-logfmt/logfmt v0.4.0 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/golang-migrate/migrate v0.0.0-20180905021119-16f2b1736e65
	github.com/golang/protobuf v1.2.0
	github.com/google/uuid v1.1.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20181017120253-0766667cb4d1 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/websocket v0.0.0-20181206070239-95ba29eb981b
	github.com/grpc-ecosystem/go-grpc-middleware v0.0.0-20190118093823-f849b5445de4
	github.com/grpc-ecosystem/go-grpc-prometheus v0.0.0-20181025070259-68e3a13e4117
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/newrelic/go-agent v2.3.0+incompatible
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.3.5
	github.com/patrickmn/go-cache v0.0.0-20180815053127-5633e0862627
	github.com/pborman/uuid v1.2.0
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v0.9.2
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
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
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.5.1
	github.com/stvp/rollbar v0.0.0-20171113052335-4a50daf855af
	go.elastic.co/apm v1.4.0
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2
	golang.org/x/oauth2 v0.0.0-20190130055435-99b60b757ec1
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	google.golang.org/api v0.1.0
	google.golang.org/genproto v0.0.0-20190128161407-8ac453e89fca // indirect
	google.golang.org/grpc v1.22.1
	gopkg.in/airbrake/gobrake.v2 v2.0.9
)

require (
	github.com/RichardKnop/redsync v1.2.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.34.0 // indirect
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20180710155616-bc664df96737 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/elastic/go-windows v1.0.0 // indirect
	github.com/fsnotify/fsnotify v1.4.7 // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.1.1 // indirect
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/gomodule/redigo v2.0.0+incompatible // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/googleapis/gax-go v2.0.0+incompatible // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.3.0 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/kelseyhightower/envconfig v1.3.0 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/kr/logfmt v0.0.0-20140226030751-b84e30acd515 // indirect
	github.com/magiconair/properties v1.8.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mongodb/mongo-go-driver v0.2.0 // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/streadway/amqp v0.0.0-20180806233856-70e15c650864 // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	go.elastic.co/fastjson v1.0.0 // indirect
	go.opencensus.io v0.18.0 // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/appengine v1.4.0 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	howett.net/plist v0.0.0-20181124034731-591f970eefbb // indirect
)
