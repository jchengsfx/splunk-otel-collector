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
	"fmt"
	"github.com/signalfx/signalfx-agent/pkg/core/config"
	"github.com/signalfx/signalfx-agent/pkg/monitors"
	"github.com/spf13/viper"
	"go.opentelemetry.io/collector/config/configmodels"
	"gopkg.in/yaml.v2"
	"reflect"
	"strings"
)

type Config struct {
	configmodels.ReceiverSettings `mapstructure:",squash"`
	// Ideally we'd be able to embed a config.MonitorConfig or derived here but because each monitor
	// has its own custom config type and field collisions (e.g. Type) prevent us from doing so.
	// custom unmarshaller with reflection to work around.
	sfxMonitorConfig interface{}
}

func (rCfg *Config) validate() error {
	if rCfg.sfxMonitorConfig == nil {
		return fmt.Errorf("must supply a valid Smart Agent Monitor config")
	}
	return nil
}

func yamlFieldsFromStructType(s reflect.Type) map[string]string {
	yamlFields := map[string]string{}
	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		tag := field.Tag
		tagYaml := strings.Split(tag.Get("yaml"), ",")[0]
		lowerTag := strings.ToLower(tagYaml)
		if tagYaml != lowerTag {
			yamlFields[lowerTag] = tagYaml
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Struct {
			otherFields := yamlFieldsFromStructType(fieldType)
			for k, v := range otherFields {
				yamlFields[k] = v
			}
		}
	}
	return yamlFields
}

func mergeConfigs(componentViperSection *viper.Viper, intoCfg interface{}) error {
	// AllSettings() will include anything not already unmarshalled in the Config instance (intoCfg).
	// This includes all Smart Agent monitor config settings, that unfortunately lost their casings in
	// viper (insensitive by design).  We need to comb through the Agent monitor config struct tags to
	// determine the expected yaml form and reset them.
	allSettings := componentViperSection.AllSettings()
	var monitorType string
	var ok bool
	if monitorType, ok = allSettings["type"].(string); !ok {
		return fmt.Errorf(`You must specify a "type" for an sfxsmartagent receiver`)
	}

	// ConfigTemplates is a map that all monitors use to register their custom configs
	// in the Smart Agent.  He we retrieve them and create a map of lowercase to expected
	// field casings
	var customMonitorConfig config.MonitorCustomConfig
	if customMonitorConfig, ok = monitors.ConfigTemplates[monitorType]; !ok {
		return fmt.Errorf("no known monitor %q\n", monitorType)
	}
	monitorConfig := reflect.TypeOf(customMonitorConfig).Elem()
	fieldCasings := yamlFieldsFromStructType(monitorConfig)

	for key, val := range allSettings {
		updatedKey := fieldCasings[key]
		if updatedKey != "" {
			delete(allSettings, key)
			allSettings[updatedKey] = val
		}
	}

	// newMonitorConfig as interface is a pointer to a monitor's config
	// (specialized one off, not MonitorCustomConfig or MonitorConfig).
	newMonitorConfig := reflect.New(monitorConfig).Interface()
	// fmt.Printf("New reflective interface instance: %#v\n\n", newMonitorConfig)
	asBytes, err := yaml.Marshal(allSettings)
	if err != nil {
		return err
	}
	// as a pointer to custom config newMonitorConfig is ok as an argumen to UnmarshalStrict()
	err = yaml.UnmarshalStrict(asBytes, newMonitorConfig)
	// fmt.Printf("instance after unmarshaling: %#v - err: %v\n", newMonitorConfig, err)
	if err != nil {
		return err
	}

	cfg := intoCfg.(*Config)
	cfg.sfxMonitorConfig = newMonitorConfig.(config.MonitorCustomConfig)
	return nil
}
