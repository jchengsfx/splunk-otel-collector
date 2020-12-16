module github.com/signalfx/splunk-otel-collector/internal/receiver/sfxsmartagentreceiver

go 1.14

require (
	github.com/signalfx/golib/v3 v3.3.16
	github.com/signalfx/signalfx-agent v1.0.1-0.20201215181116-ee48f3743d6a
	github.com/stretchr/testify v1.6.1
	go.opentelemetry.io/collector v0.15.0
	go.uber.org/zap v1.16.0
)

replace (
	code.cloudfoundry.org/go-loggregator => github.com/signalfx/go-loggregator v1.0.1-0.20200205155641-5ba5ca92118d
	github.com/dancannon/gorethink => gopkg.in/gorethink/gorethink.v4 v4.0.0
	github.com/influxdata/telegraf => github.com/signalfx/telegraf v0.10.2-0.20201211214327-200738592ced
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20201105135750-00f16d1ac3a4
	github.com/signalfx/signalfx-agent => ../../../../signalfx-agent
	github.com/signalfx/signalfx-agent/pkg/apm => ../../../../signalfx-agent/pkg/apm
	github.com/signalfx/splunk-otel-collector/internal/receiver/sfxsmartagentreceiver v0.0.0-00010101000000-000000000000 => ./internal/receiver/sfxsmartagentreceiver
	github.com/soheilhy/cmux => ../../../../signalfx-agent/thirdparty/cmux
	google.golang.org/grpc => google.golang.org/grpc v1.29.1
)
