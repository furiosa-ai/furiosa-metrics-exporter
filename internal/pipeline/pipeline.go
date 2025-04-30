package pipeline

import (
	"sync"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/collector"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
)

type Pipeline struct {
	collectors []collector.Collector
}

func NewRegisteredPipeline(devices []smi.Device, metricFactory collector.MetricFactory, kubeResMapper collector.KubeResourcesMapper) *Pipeline {
	p := Pipeline{
		collectors: []collector.Collector{
			collector.NewTemperatureCollector(devices, metricFactory, kubeResMapper),
			collector.NewPowerCollector(devices, metricFactory, kubeResMapper),
			collector.NewLivenessCollector(devices, metricFactory, kubeResMapper),
			collector.NewCoreUtilizationCollector(devices, metricFactory, kubeResMapper),
			collector.NewCoreFrequencyCollector(devices, metricFactory, kubeResMapper),
			collector.NewCycleCollector(devices, metricFactory, kubeResMapper),
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
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := p.collectors[i].Collect(); err != nil {
				errors[i] = err
			}
		}()
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
