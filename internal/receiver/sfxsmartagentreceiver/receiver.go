// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sfxsmartagentreceiver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/signalfx/signalfx-agent/pkg/core/common/constants"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/core/meta"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/genericjmx"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/kafka"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/python"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/redis"
	"github.com/signalfx/signalfx-agent/pkg/monitors/collectd/uptime"
	"github.com/signalfx/signalfx-agent/pkg/monitors/cpu"
	"github.com/signalfx/signalfx-agent/pkg/monitors/subproc"
	"github.com/signalfx/signalfx-agent/pkg/monitors/telegraf/monitors/procstat"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
	"github.com/signalfx/signalfx-agent/pkg/utils"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

// sfxSmartAgentReceiver implements the component.MetricsReceiver for SignalFx metric protocol.
type sfxSmartAgentReceiver struct {
	sync.Mutex
	logger       *zap.Logger
	config       *Config
	monitor      *interface{}
	output       *Output
	nextConsumer consumer.MetricsConsumer
	server       *http.Server

	startOnce sync.Once
	stopOnce  sync.Once
}

var _ component.MetricsReceiver = (*sfxSmartAgentReceiver)(nil)

func NewReceiver(
	logger *zap.Logger,
	config Config,
	nextConsumer consumer.MetricsConsumer,
) *sfxSmartAgentReceiver {
	r := &sfxSmartAgentReceiver{
		logger: logger,
		config: &config,
		nextConsumer: nextConsumer,
	}
	return r
}

func (r *sfxSmartAgentReceiver) Start(_ context.Context, host component.Host) error {
	r.Lock()
	defer r.Unlock()

	agentMeta := &meta.AgentMeta{
		InternalStatusHost: "0.0.0.",
		InternalStatusPort: 12345,
	}

	// These are required for collectd and sfxcollectd monitor types
	manager := monitors.NewMonitorManager(agentMeta)
	collectdManager(manager)

	monitorConfig := r.config.sfxMonitorConfig
	monitorConfigCore := monitorConfig.(config.MonitorCustomConfig).MonitorConfigCore()
	monitorType := monitorConfigCore.Type
	monitorName := strings.Replace(r.config.Name(), "/", "", -1)
	monitorConfigCore.MonitorID = types.MonitorID(monitorName)
	monitorFactory := monitors.MonitorFactories[monitorType]
	monitor := monitorFactory()
	r.monitor = &monitor

	output := &Output{nextConsumer: r.nextConsumer}
	r.output = output

	// Taken from signalfx-agent activemonitor.  Should be exported in that lib in future.
	outputValue := utils.FindFieldWithEmbeddedStructs(monitor, "Output",
		reflect.TypeOf((*types.Output)(nil)).Elem())
	if !outputValue.IsValid() {
		outputValue = utils.FindFieldWithEmbeddedStructs(monitor, "Output",
			reflect.TypeOf((*types.FilteringOutput)(nil)).Elem())
		if !outputValue.IsValid() {
			return fmt.Errorf("invalid monitor instance: %#v", monitor)
		}
	}
	outputValue.Set(reflect.ValueOf(output))

	return config.CallConfigure(monitor, monitorConfig)
}

func (r *sfxSmartAgentReceiver) Shutdown(context.Context) error {
	return nil
}

func collectdManager(monitorManager *monitors.MonitorManager) *collectd.Manager {
	collectdConfig := &config.CollectdConfig{
		DisableCollectd:      false,
		Timeout:              40,
		ReadThreads:          5,
		WriteThreads:         2,
		WriteQueueLimitHigh:  500000,
		WriteQueueLimitLow:   400000,
		LogLevel:             "notice",
		IntervalSeconds:      10,
		WriteServerIPAddr:    "127.9.8.7",
		WriteServerPort:      0,
		ConfigDir:            "/etc/signalfx",
		BundleDir:            os.Getenv(constants.BundleDirEnvVar),
		HasGenericJMXMonitor: true,
		// InstanceName string `yaml:"-"`
		// WriteServerQuery string `yaml:"-"`
	}
	monitorConfigs := []config.MonitorConfig{}
	fmt.Printf("Starting monitorManager w/ collectd singleton\n")
	monitorManager.Configure(monitorConfigs, collectdConfig, 10)
	collectdManager := collectd.MainInstance()
	fmt.Printf("Started collectd singleton %#v\n", collectdManager)
	return collectdManager
}

// sfxcollectd
func redisMonitor() error {
	fmt.Printf("Creating Redis monitor\n")
	redisMonitor := redis.Monitor{
		python.PyMonitor{
			MonitorCore: subproc.New(),
			Output:      &Output{},
		},
	}
	fmt.Printf("Created Redis monitor: %#v\n", redisMonitor)

	fmt.Printf("Creating Redis config\n")
	redisConfig := redis.Config{
		MonitorConfig: config.MonitorConfig{IntervalSeconds: 5},
		Host:          "localhost", Port: 6379,
	}
	fmt.Printf("Created Redis config: %#v\n", redisConfig)

	fmt.Printf("Configuring Redis monitor\n")
	err := redisMonitor.Configure(&redisConfig)
	fmt.Printf("Configured Redis monitor. err: %#v\n", err)
	return err
}

// collectd/GenericJMX
func kafkaMonitor(collectdManager *collectd.Manager) error {
	monitorID := types.MonitorID("__soc_kafka__")

	fmt.Printf("Creating Kafka monitor\n")
	kafkaMonitor := kafka.Monitor{
		JMXMonitorCore: genericjmx.NewJMXMonitorCore(kafka.DefaultMBeans(), "kafka"),
	}
	output := &Output{}
	kafkaMonitor.JMXMonitorCore.MonitorCore.Output = output
	fmt.Printf("Created Kafka monitor: %#v\n", kafkaMonitor)

	fmt.Printf("Creating Kafka config\n")
	kafkaConfig := kafka.Config{
		Config: genericjmx.Config{
			MonitorConfig: config.MonitorConfig{
				IntervalSeconds: 5,
				MonitorID:       monitorID,
			},
			ServiceURL: "service:jmx:rmi:///jndi/rmi://localhost:7199/jmxrmi",
		},
	}
	fmt.Printf("Created Kafka config: %#v\n", kafkaConfig)

	fmt.Printf("Configuring Kafka monitor\n")
	err := kafkaMonitor.Configure(&kafkaConfig)
	fmt.Printf("Configured Kafka monitor. err: %#v\n", err)
	err = collectdManager.ConfigureFromMonitor(monitorID, output, true)
	fmt.Printf("Ran collectdManager.ConfigureFromMonitor. err: %#v\n", err)
	return err
}

func uptimeMonitor(collectdManager *collectd.Manager) error {
	monitorID := types.MonitorID("__soc_uptime__")
	fmt.Printf("Creating collectd/uptime monitor\n")
	uptimeMonitor := uptime.Monitor{
		MonitorCore: *collectd.NewMonitorCore(uptime.CollectdTemplate),
	}
	fmt.Printf("Created collectd/uptime monitor: %#v\n", uptimeMonitor)

	output := &Output{}
	uptimeMonitor.MonitorCore.Output = output

	fmt.Printf("Creating collectd/update config\n")
	uptimeConfig := uptime.Config{
		MonitorConfig: config.MonitorConfig{
			MonitorID: monitorID,
		},
	}
	fmt.Printf("Created collectd/update config: %#v\n", uptimeConfig)

	fmt.Printf("Configuring collectd/update monitor\n")
	err := uptimeMonitor.Configure(&uptimeConfig)
	fmt.Printf("Configured collectd/update monitor. err: %#v\n", err)

	err = collectdManager.ConfigureFromMonitor(monitorID, output, false)
	fmt.Printf("Ran collectdManager.ConfigureFromMonitor. err: %#v\n", err)
	return err
}

func telegrafProcstatMonitor() error {
	monitorID := types.MonitorID("__soc_telegraf_procstat__")
	fmt.Printf("Creating Procstat monitor\n")
	procstatMonitor := procstat.Monitor{
		Output: &Output{},
	}
	fmt.Printf("Created Procstat monitor: %#v\n", procstatMonitor)

	fmt.Printf("Creating Procstat config\n")
	procstatConfig := procstat.Config{
		MonitorConfig: config.MonitorConfig{
			IntervalSeconds: 5,
			MonitorID:       monitorID,
		},
		Exe: ".*otelcol.*",
	}
	fmt.Printf("Created Procstat config: %#v\n", procstatConfig)

	fmt.Printf("Configuring Procstat monitor\n")
	err := procstatMonitor.Configure(&procstatConfig)
	fmt.Printf("Configured Procstat monitor. err: %#v\n", err)
	return err
}

func cpuMonitor() error {
	monitorID := types.MonitorID("__soc_cpu__")
	fmt.Printf("Creating CPU monitor\n")
	cpuMonitor := cpu.Monitor{
		Output: &Output{},
	}
	fmt.Printf("Created CPU monitor: %#v\n", cpuMonitor)

	fmt.Printf("Creating CPU config\n")
	cpuConfig := cpu.Config{
		MonitorConfig: config.MonitorConfig{
			IntervalSeconds: 1,
			MonitorID:       monitorID,
		},
		ReportPerCPU: false,
	}
	fmt.Printf("Created CPU config: %#v\n", cpuConfig)

	fmt.Printf("Configuring CPU monitor\n")
	err := cpuMonitor.Configure(&cpuConfig)
	fmt.Printf("Configured CPU monitor. err: %#v\n", err)
	return err
}
