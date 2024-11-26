package pipeline

import (
	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/collector"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
)

type Pipeline struct {
	collectors []collector.Collector
}

func NewRegisteredPipeline(devices []smi.Device, nodeName string) *Pipeline {
	p := Pipeline{
		collectors: []collector.Collector{
			collector.NewTemperatureCollector(devices, nodeName),
			collector.NewPowerCollector(devices, nodeName),
			collector.NewLivenessCollector(devices, nodeName),
			collector.NewErrorCollector(devices, nodeName),
			collector.NewCoreUtilizationCollector(devices, nodeName),
			//collector.NewMemoryCollector(devices, nodeName),
		},
	}

	for _, c := range p.collectors {
		c.Register()
	}

	return &p
}

func (p *Pipeline) Collect() []error {
	errors := make([]error, 0)
	for _, c := range p.collectors {
		if err := c.Collect(); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
