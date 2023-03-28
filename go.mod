module github.com/carousell/Orion

go 1.17

replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
	github.com/openzipkin-contrib/zipkin-go-opentracing => github.com/openzipkin-contrib/zipkin-go-opentracing v0.3.5
)

require (
	github.com/DataDog/datadog-go v0.0.0-20180822151419-281ae9f2d895 // indirect
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
	github.com/frankban/quicktest v1.13.0 // indirect
	github.com/getsentry/raven-go v0.0.0-20190125112653-238ebd86338d
	github.com/go-kit/kit v0.8.0
	github.com/go-logfmt/logfmt v0.4.0 // indirect
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
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
	github.com/pborman/uuid v1.2.0
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.2
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90 // indirect
	github.com/prometheus/common v0.2.0 // indirect
	github.com/prometheus/procfs v0.0.3 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20181016184325-3113b8401b8a // indirect
	github.com/santhosh-tekuri/jsonschema v1.2.4 // indirect
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/afero v1.2.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.4.0
	github.com/stvp/rollbar v0.0.0-20171113052335-4a50daf855af
	go.elastic.co/apm v1.4.0
	golang.org/x/crypto v0.0.0-20220411220226-7b82a4e95df4 // indirect
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f // indirect
	google.golang.org/genproto v0.0.0-20190128161407-8ac453e89fca // indirect
	google.golang.org/grpc v1.22.1
	gopkg.in/airbrake/gobrake.v2 v2.0.9
)

require (
	github.com/Shopify/sarama v1.19.0 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/elastic/go-windows v1.0.0 // indirect
	github.com/fsnotify/fsnotify v1.4.7 // indirect
	github.com/gogo/protobuf v1.1.1 // indirect
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hpcloud/tail v1.0.0 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/jtolds/gls v4.2.1+incompatible // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.1 // indirect
	github.com/kr/logfmt v0.0.0-20140226030751-b84e30acd515 // indirect
	github.com/magiconair/properties v1.8.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/onsi/ginkgo v1.6.0 // indirect
	github.com/onsi/gomega v1.4.1 // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/smartystreets/assertions v0.0.0-20180820201707-7c9eb446e3cf // indirect
	github.com/smartystreets/goconvey v0.0.0-20180222194500-ef6db91d284a // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	go.elastic.co/fastjson v1.0.0 // indirect
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1 // indirect
	golang.org/x/text v0.3.8 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	howett.net/plist v0.0.0-20181124034731-591f970eefbb // indirect
)
