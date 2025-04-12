package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	model "github.com/prometheus/client_model/go"
)

type LabelFilterCollector struct {
	collector  prometheus.Collector
	opts       prometheus.Opts
	metricType prometheus.ValueType
}

func NewLabelFilterCollector(collector prometheus.Collector, opts prometheus.Opts, metricType prometheus.ValueType) prometheus.Collector {
	return &LabelFilterCollector{
		collector:  collector,
		opts:       opts,
		metricType: metricType,
	}
}

func (c *LabelFilterCollector) Describe(ch chan<- *prometheus.Desc) {
	c.collector.Describe(ch)
}

func (c *LabelFilterCollector) Collect(ch chan<- prometheus.Metric) {
	metricChan := make(chan prometheus.Metric)
	go func() {
		c.collector.Collect(metricChan)
		close(metricChan)
	}()

	for metric := range metricChan {
		m := &model.Metric{}
		if err := metric.Write(m); err != nil {
			continue
		}

		filteredLabels := filterEmptyLabels(m.Label)
		newDesc := prometheus.NewDesc(
			prometheus.BuildFQName(c.opts.Namespace, c.opts.Subsystem, c.opts.Name),
			c.opts.Help,
			filteredLabels.keys,
			c.opts.ConstLabels,
		)

		var value float64

		switch c.metricType {
		case prometheus.GaugeValue:
			value = m.GetGauge().GetValue()

		case prometheus.CounterValue:
			value = m.GetCounter().GetValue()

		default:
			continue
		}

		newMetric, err := prometheus.NewConstMetric(
			newDesc,
			c.metricType,
			value,
			filteredLabels.values...,
		)

		if err == nil {
			ch <- newMetric
		}
	}
}

type filteredLabels struct {
	keys   []string
	values []string
}

func filterEmptyLabels(labelPairs []*model.LabelPair) filteredLabels {
	result := filteredLabels{
		keys:   []string{},
		values: []string{},
	}

	for _, pair := range labelPairs {
		if pair.GetValue() != "" {
			result.keys = append(result.keys, pair.GetName())
			result.values = append(result.values, pair.GetValue())
		}
	}

	return result
}
