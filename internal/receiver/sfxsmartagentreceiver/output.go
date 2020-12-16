package sfxsmartagentreceiver

import (
	"fmt"

	"github.com/signalfx/golib/v3/datapoint"
	"github.com/signalfx/golib/v3/event"
	"github.com/signalfx/golib/v3/trace"
	"github.com/signalfx/signalfx-agent/pkg/core/dpfilters"
	"github.com/signalfx/signalfx-agent/pkg/monitors/types"
)

type Output struct{}

var _ types.FilteringOutput = (*Output)(nil)

func (output *Output) AddDatapointExclusionFilter(filter dpfilters.DatapointFilter) {
	fmt.Printf("AddDatapointExclusionFilter: %#v\n", filter)
}

func (output *Output) EnabledMetrics() []string {
	fmt.Printf("EnabledMetrics\n")
	return []string{}
}

func (output *Output) HasEnabledMetricInGroup(group string) bool {
	fmt.Printf("HasEnabledMetricInGroup: %v\n", group)
	return false
}

func (output *Output) HasAnyExtraMetrics() bool {
	fmt.Printf("HasAnyExtraMetrics\n")
	return false
}

func (output *Output) Copy() types.Output {
	fmt.Printf("Copy\n")
	return output
}

func (output *Output) SendDatapoints(datapoint ...*datapoint.Datapoint) {
	fmt.Printf("SendDatapoints: %v datapoints.\n", len(datapoint))
	for _, i := range datapoint {
		fmt.Printf("SendDatapoint: %#v\n", i)
	}
}

func (output *Output) SendEvent(event *event.Event) {
	fmt.Printf("SendEvent: %#v\n", event)
}

func (output *Output) SendSpans(span ...*trace.Span) {
	fmt.Printf("SendSpans: %#v\n", span)
}

func (output *Output) SendDimensionUpdate(dimension *types.Dimension) {
	fmt.Printf("SendDimensionUpdate: %#v\n", dimension)
}

func (output *Output) AddExtraDimension(key, value string) {
	fmt.Printf("AddExtraDimension: %#v - %#v\n", key, value)
}

func (output *Output) RemoveExtraDimension(key string) {
	fmt.Printf("RemoveExtraDimension: %#v\n", key)
}

func (output *Output) AddExtraSpanTag(key, value string) {
	fmt.Printf("AddExtraSpanTag: %#v - %#v\n", key, value)
}

func (output *Output) RemoveExtraSpanTag(key string) {
	fmt.Printf("RemoveExtraSpanTag: %#v\n", key)
}

func (output *Output) AddDefaultSpanTag(key, value string) {
	fmt.Printf("AddDefaultSpanTag: %#v - %#v\n", key, value)
}

func (output *Output) RemoveDefaultSpanTag(key string) {
	fmt.Printf("RemoveDefaultSpanTag: %#v\n", key)
}
