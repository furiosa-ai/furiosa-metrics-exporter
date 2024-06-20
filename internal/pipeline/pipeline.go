package pipeline

import (
	"github.com/furiosa-ai/furiosa-metric-exporter/internal/collector"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/smi"
)

type Pipeline struct {
	Collectors []collector.Collector
}

func NewRegisteredPipeline(devices []smi.Device) *Pipeline {
	p := Pipeline{
		Collectors: []collector.Collector{
			collector.NewTemperatureCollector(devices),
			//TODO: add more collectors
		},
	}

	for _, c := range p.Collectors {
		c.Register()
	}

	return &p
}

func (p *Pipeline) Collect() error {
	for _, c := range p.Collectors {
		err := c.Collect()
		if err != nil {
			return err
		}
	}

	return nil
}
