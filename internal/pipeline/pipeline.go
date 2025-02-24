package pipeline

import (
	"net/http"
	"sync"

	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/collector"
	"github.com/furiosa-ai/furiosa-metrics-exporter/internal/kubernetes"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Pipeline struct {
	collectors []collector.Collector
}

func NewRegisteredPipeline(devices []smi.Device, nodeName string, registryWithPod *prometheus.Registry) *Pipeline {
	p := Pipeline{
		collectors: []collector.Collector{
			collector.NewTemperatureCollector(devices, nodeName),
			collector.NewPowerCollector(devices, nodeName),
			collector.NewLivenessCollector(devices, nodeName),
			collector.NewCoreUtilizationCollector(devices, nodeName),
			collector.NewCoreFrequencyCollector(devices, nodeName),
			//collector.NewMemoryCollector(devices, nodeName),
		},
	}

	http.Handle("/metrics", promhttp.HandlerFor(prometheus.Gatherers{
		prometheus.DefaultGatherer,
		registryWithPod,
	}, promhttp.HandlerOpts{}))

	for _, c := range p.collectors {
		c.Register(registryWithPod)
	}

	return &p
}

func (p *Pipeline) Collect(devicePodMap map[string]kubernetes.PodInfo) []error {
	errors := make([]error, len(p.collectors))

	wg := new(sync.WaitGroup)
	for i := range p.collectors {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := p.collectors[i].Collect(devicePodMap); err != nil {
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
