package pipeline

import (
	"sync"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/collector"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
)

type Pipeline struct {
	collectors []collector.Collector
}

func NewRegisteredPipeline(devices []smi.Device, metricFactory collector.MetricFactory) *Pipeline {
	p := Pipeline{
		collectors: []collector.Collector{
			collector.NewTemperatureCollector(devices, metricFactory),
			collector.NewPowerCollector(devices, metricFactory),
			collector.NewLivenessCollector(devices, metricFactory),
			collector.NewCoreUtilizationCollector(devices, metricFactory),
			collector.NewCoreFrequencyCollector(devices, metricFactory),
			collector.NewTotalCycleCountCollector(devices, metricFactory),
			collector.NewTaskExecutionCycleCollector(devices, metricFactory),
			//collector.NewMemoryCollector(devices, nodeName),
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
