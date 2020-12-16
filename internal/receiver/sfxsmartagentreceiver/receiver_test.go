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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.uber.org/zap"
)

func TestSfxSmartAgentReceiver(t *testing.T) {
	defaultConfig := CreateDefaultConfig().(*Config)
	type args struct {
		config       Config
		nextConsumer consumer.MetricsConsumer
	}
	tests := []struct {
		name         string
		args         args
	}{
		{
			name: "default_endpoint",
			args: args{
				config:       *defaultConfig,
				nextConsumer: consumertest.NewMetricsNop(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewReceiver(zap.NewNop(), tt.args.config)
			if tt.args.nextConsumer != nil {
				got.NextConsumer = tt.args.nextConsumer
			}
			err := got.Start(context.Background(), componenttest.NewNopHost())
			assert.NoError(t, err)
		})
	}
}
