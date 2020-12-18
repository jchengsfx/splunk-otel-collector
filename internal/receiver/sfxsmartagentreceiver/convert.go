package sfxsmartagentreceiver

import (
	"fmt"
	sfxdatapoint "github.com/signalfx/golib/v3/datapoint"
	"go.opentelemetry.io/collector/consumer/pdata"
	"go.uber.org/zap"
)


var (
	logger zap.Logger
	errUnsupportedMetricTypeTimestamp = fmt.Errorf("unsupported metric type timestamp")
	errUnsupportedValueTypeTimestamp = fmt.Errorf("unsupported value type string")
	errNoIntValue = fmt.Errorf("no valid value for expected IntValue")
	errNoFloatValue = fmt.Errorf("no valid value for expected FloatValue")
)

// Based on https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/v0.15.0/receiver/signalfxreceiver/signalfxv2_to_metricdata.go
func datapointsToMetrics(datapoints []*sfxdatapoint.Datapoint) (pdata.Metrics, int) {
	numDropped := 0
	md := pdata.NewMetrics()
	md.ResourceMetrics().Resize(1)
	rm := md.ResourceMetrics().At(0)

	rm.InstrumentationLibraryMetrics().Resize(1)
	ilm := rm.InstrumentationLibraryMetrics().At(0)

	metrics := ilm.Metrics()
	metrics.Resize(len(datapoints))

	i := 0
	for _, datapoint := range datapoints {
		if datapoint == nil {
			continue
		}

		m := metrics.At(i)
		// First check if the type is convertible and the data point is consistent.
		err := fillInType(datapoint, m)
		if err != nil {
			numDropped++
			logger.Debug("SignalFx data-point type conversion error",
				zap.Error(err),
				zap.String("metric", datapoint.String()))
			continue
		}

		m.SetName(datapoint.Metric)

		switch m.DataType() {
		case pdata.MetricDataTypeIntGauge:
			err = fillIntDatapoint(datapoint, m.IntGauge().DataPoints())
		case pdata.MetricDataTypeIntSum:
			err = fillIntDatapoint(datapoint, m.IntSum().DataPoints())
		case pdata.MetricDataTypeDoubleGauge:
			err = fillDoubleDatapoint(datapoint, m.DoubleGauge().DataPoints())
		case pdata.MetricDataTypeDoubleSum:
			err = fillDoubleDatapoint(datapoint, m.DoubleSum().DataPoints())
		}

		if err != nil {
			numDropped++
			logger.Debug("SignalFx data-point datum conversion error",
				zap.Error(err),
				zap.String("metric", datapoint.Metric))
			continue
		}

		i++
	}

	metrics.Resize(i)

	return md, numDropped


}
func fillInType(datapoint *sfxdatapoint.Datapoint, m pdata.Metric) error {
	sfxMetricType := datapoint.MetricType
	if sfxMetricType == sfxdatapoint.Timestamp {
		return errUnsupportedMetricTypeTimestamp
	}

	var isFloat bool
	switch datapoint.Value.(type) {
	case sfxdatapoint.IntValue:
		isFloat = false
	case sfxdatapoint.FloatValue:
		isFloat = true
	case sfxdatapoint.StringValue:
		return errUnsupportedMetricTypeTimestamp
	}

	switch sfxMetricType {
	case sfxdatapoint.Gauge,sfxdatapoint.Enum, sfxdatapoint.Rate, sfxdatapoint.Timestamp:
		if isFloat {
			m.SetDataType(pdata.MetricDataTypeDoubleGauge)
			m.DoubleGauge().InitEmpty() // will need to be removed w/ 0.16.0 adoption
		} else {
			m.SetDataType(pdata.MetricDataTypeIntGauge)
			m.IntGauge().InitEmpty() // will need to be removed w/ 0.16.0 adoption
		}
	case sfxdatapoint.Count:
		if isFloat {
			m.SetDataType(pdata.MetricDataTypeDoubleSum)
			m.DoubleSum().InitEmpty() // will need to be removed w/ 0.16.0 adoption
			m.DoubleSum().SetAggregationTemporality(pdata.AggregationTemporalityDelta)
			m.DoubleSum().SetIsMonotonic(true)
		} else {
			m.SetDataType(pdata.MetricDataTypeIntSum)
			m.IntSum().InitEmpty() // will need to be removed w/ 0.16.0 adoption
			m.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityDelta)
			m.IntSum().SetIsMonotonic(true)
		}
	case sfxdatapoint.Counter:
		if isFloat {
			m.SetDataType(pdata.MetricDataTypeDoubleSum)
			m.DoubleSum().InitEmpty() // will need to be removed w/ 0.16.0 adoption
			m.DoubleSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
			m.DoubleSum().SetIsMonotonic(true)
		} else {
			m.SetDataType(pdata.MetricDataTypeIntSum)
			m.IntSum().InitEmpty() // will need to be removed w/ 0.16.0 adoption
			m.IntSum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
			m.IntSum().SetIsMonotonic(true)
		}
	default:
		return fmt.Errorf("unknown datapoint metric type %v", sfxMetricType)
	}

	return nil
}

func fillIntDatapoint(datapoint *sfxdatapoint.Datapoint, dps pdata.IntDataPointSlice) error {
	var intValue sfxdatapoint.IntValue
	var ok bool
	if intValue, ok = datapoint.Value.(sfxdatapoint.IntValue); !ok {
		return errNoIntValue
	}

	dps.Resize(1)
	dp := dps.At(0)
	dp.SetTimestamp(pdata.TimestampUnixNano(uint64(datapoint.Timestamp.UnixNano())))
	dp.SetValue(intValue.Int())
	fillInLabels(datapoint.Dimensions, dp.LabelsMap())

	return nil
}

func fillDoubleDatapoint(datapoint *sfxdatapoint.Datapoint, dps pdata.DoubleDataPointSlice) error {
	var floatValue sfxdatapoint.FloatValue
	var ok bool
	if floatValue, ok = datapoint.Value.(sfxdatapoint.FloatValue); !ok {
		return errNoFloatValue
	}

	dps.Resize(1)
	dp := dps.At(0)
	dp.SetTimestamp(pdata.TimestampUnixNano(uint64(datapoint.Timestamp.UnixNano())))
	dp.SetValue(floatValue.Float())
	fillInLabels(datapoint.Dimensions, dp.LabelsMap())

	return nil

}

func fillInLabels(dimensions map[string]string, labels pdata.StringMap) {
	labels.InitEmptyWithCapacity(len(dimensions))
	for k, v := range dimensions {
		labels.Insert(k, v)
	}
}
