package pipeline

import (
	"sync"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/collector"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

type Pipeline struct {
	collectors []collector.Collector
}

func NewRegisteredPipeline(deviceSource collector.DeviceSource, sampler collector.CoreUtilizationSampler, metricFactory collector.MetricFactory, kubeResMapper collector.KubeResourcesMapper) *Pipeline {
	p := Pipeline{
		collectors: []collector.Collector{
			collector.NewTemperatureCollector(deviceSource, metricFactory, kubeResMapper),
			collector.NewPowerCollector(deviceSource, metricFactory, kubeResMapper),
			collector.NewLivenessCollector(deviceSource, metricFactory, kubeResMapper),
			collector.NewCoreUtilizationCollector(deviceSource, sampler, metricFactory, kubeResMapper),
			collector.NewCoreFrequencyCollector(deviceSource, metricFactory, kubeResMapper),
			collector.NewCycleCollector(deviceSource, metricFactory, kubeResMapper),
			collector.NewThrottleReasonCollector(deviceSource, metricFactory, kubeResMapper),
			collector.NewMemoryCollector(deviceSource, metricFactory, kubeResMapper),
		},
	}

	for _, c := range p.collectors {
		c.Register()
	}

	return &p
}

func (p *Pipeline) Collect() []error {
	errors := make([]error, len(p.collectors))

	wg := new(sync.WaitGroup)
	for i := range p.collectors {
		wg.Go(func() {
			c := p.collectors[i]

			metrics := make(map[*api.Device]collector.Metric, 0)

			for _, d := range c.GetDevices() {
				metric, err := c.GetMetricFactory().NewDeviceWiseMetric(d)
				if err != nil {
					errors[i] = err
					continue
				}
				metrics[d] = metric
			}

			if err := c.Collect(metrics); err != nil {
				errors[i] = err
			}
		})
	}
	wg.Wait()

	results := make([]error, 0)
	for i := range errors {
		if errors[i] != nil {
			results = append(results, errors[i])
		}
	}
	return results
}
