package provider

import (
	"errors"
	"fmt"

	prom "github.com/directxman12/k8s-prometheus-adapter/pkg/client"
	"github.com/prometheus/common/model"
	"k8s.io/metrics/pkg/apis/external_metrics"
)

type vectorConverter struct {
	SampleConverter SampleConverter
}

//NewVectorConverter creates a VectorConverter capable of converting
//vector Prometheus query results into external metric types.
func NewVectorConverter(sampleConverter *SampleConverter) MetricConverter {
	return &vectorConverter{
		SampleConverter: *sampleConverter,
	}
}

func (c *vectorConverter) Convert(queryResult prom.QueryResult) (*external_metrics.ExternalMetricValueList, error) {
	if queryResult.Type != model.ValVector {
		return nil, errors.New("vectorConverter can only convert scalar query results")
	}

	toConvert := *queryResult.Vector

	if toConvert == nil {
		return nil, errors.New("the provided input did not contain vector query results")
	}

	return c.convert(toConvert)
}

func (c *vectorConverter) convert(result model.Vector) (*external_metrics.ExternalMetricValueList, error) {
	items := []external_metrics.ExternalMetricValue{}
	metricValueList := external_metrics.ExternalMetricValueList{
		Items: items,
	}

	numSamples := result.Len()
	if numSamples == 0 {
		return &metricValueList, nil
	}

	for _, val := range result {

		singleMetric, err := c.SampleConverter.Convert(val)

		if err != nil {
			return nil, fmt.Errorf("unable to convert vector: %v", err)
		}

		items = append(items, *singleMetric)
	}

	metricValueList = external_metrics.ExternalMetricValueList{
		Items: items,
	}
	return &metricValueList, nil
}
