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

package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.uber.org/zap"

	sfxa "github.com/signalfx/splunk-otel-collector/internal/receiver/sfxsmartagentreceiver"
)

func main() {
	config := sfxa.CreateDefaultConfig().(*sfxa.Config)
	receiver := sfxa.NewReceiver(zap.NewNop(), *config)
	receiver.NextConsumer = consumertest.NewMetricsNop()
	err := receiver.Start(context.Background(), componenttest.NewNopHost())
	fmt.Printf("FINISHED: %v\n", err)
}
