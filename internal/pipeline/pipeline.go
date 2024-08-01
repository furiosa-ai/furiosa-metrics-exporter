package pipeline

import (
	"context"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/collector"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/smi"
	"golang.org/x/sync/errgroup"
)

type Pipeline struct {
	Collectors []collector.Collector
}

func NewRegisteredPipeline(devices []smi.Device) *Pipeline {
	p := Pipeline{
		Collectors: []collector.Collector{
			collector.NewTemperatureCollector(devices),
			collector.NewPowerCollector(devices),
			collector.NewLivenessCollector(devices),
			collector.NewErrorCollector(devices),
			//collector.NewUtilizationCollector(devices),
			//collector.NewMemoryCollector(devices),
		},
	}

	for _, c := range p.Collectors {
		c.Register()
	}

	return &p
}

func (p *Pipeline) Collect() error {
	errGroup, _ := errgroup.WithContext(context.TODO())
	for _, c := range p.Collectors {
		c := c
		errGroup.Go(func() error {
			return c.Collect()
		})
	}

	if err := errGroup.Wait(); err != nil {
		return err
	}

	return nil
}
