package pipeline

import (
	"sync"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/collector"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
)

type Pipeline struct {
	devices    []smi.Device
	collectors []collector.Collector
}

func NewRegisteredPipeline(devices []smi.Device, metricFactory collector.MetricFactory) *Pipeline {
	deviceNames := make([]string, 0, len(devices))
	for _, d := range devices {
		deviceInfo, err := d.DeviceInfo()
		if err != nil {
			continue
		}

		deviceNames = append(deviceNames, deviceInfo.Name())
	}

	p := Pipeline{
		devices: devices,
		collectors: []collector.Collector{
			collector.NewTemperatureCollector(deviceNames, metricFactory),
			collector.NewPowerCollector(deviceNames, metricFactory),
			collector.NewLivenessCollector(deviceNames, metricFactory),
			collector.NewCoreUtilizationCollector(deviceNames, metricFactory),
			collector.NewCoreFrequencyCollector(deviceNames, metricFactory),
			collector.NewTotalCycleCountCollector(deviceNames, metricFactory),
			collector.NewTaskExecutionCycleCollector(deviceNames, metricFactory),
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

	collector.SyncDeviceSMICache(p.devices)
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
