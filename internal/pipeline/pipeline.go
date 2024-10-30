package pipeline

import (
	"context"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/collector"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"golang.org/x/sync/errgroup"
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

func (p *Pipeline) Collect() error {
	errGroup, _ := errgroup.WithContext(context.TODO())

	for i := range p.collectors {
		errGroup.Go(func() error {
			return p.collectors[i].Collect()
		})
	}

	if err := errGroup.Wait(); err != nil {
		return err
	}

	return nil
}
