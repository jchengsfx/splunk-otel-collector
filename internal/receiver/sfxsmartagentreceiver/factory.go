// Copyright 2019, OpenTelemetry Authors
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
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
)

const (
	typeStr = "sfxsmartagent"
)

func NewFactory() component.ReceiverFactory {
	return receiverhelper.NewFactory(
		typeStr,
		CreateDefaultConfig,
		receiverhelper.WithMetrics(createMetricsReceiver),
	)
}

func CreateDefaultConfig() configmodels.Receiver {
	return &Config{
		ReceiverSettings: configmodels.ReceiverSettings{
			TypeVal: typeStr,
			NameVal: typeStr,
		},
	}
}

func (rCfg *Config) validate() error {
	return nil
}

func createMetricsReceiver(
	_ context.Context,
	params component.ReceiverCreateParams,
	cfg configmodels.Receiver,
	consumer consumer.MetricsConsumer,
) (component.MetricsReceiver, error) {
	rCfg := cfg.(*Config)

	err := rCfg.validate()
	if err != nil {
		return nil, err
	}

	receiverLock.Lock()
	r := receivers[rCfg]
	if r == nil {
		r = NewReceiver(params.Logger, *rCfg)
		receivers[rCfg] = r
	}
	receiverLock.Unlock()

	r.NextConsumer = consumer

	return r, nil
}

var receiverLock sync.Mutex
var receivers = map[*Config]*sfxSmartAgentReceiver{}
