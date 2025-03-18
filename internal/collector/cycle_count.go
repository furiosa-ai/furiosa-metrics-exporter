package collector

import (
	"errors"
	"fmt"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	cycleCount = "cycle_count"
)

type cycleCountCollector struct {
	devices    []smi.Device
	counterVec *prometheus.CounterVec
	nodeName   string
}

var _ Collector = (*cycleCountCollector)(nil)

var (
	previousCycleCountMap map[uint32]float64 = make(map[uint32]float64)
)

func NewCycleCountCollector(devices []smi.Device, nodeName string) Collector {
	return &cycleCountCollector{
		devices:  devices,
		nodeName: nodeName,
	}
}

func (t *cycleCountCollector) Register() {
	t.counterVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "furiosa_npu_cycle_count",
		Help: "The current cycle count of NPU device",
	},
		[]string{
			arch,
			device,
			core,
			kubernetesNodeName,
			uuid,
		})
}

func (t *cycleCountCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		info, err := getDeviceInfo(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		metric_map := make(map[uint32]Metric)

		cores := info.cores
		for _, core_index := range cores {
			metric := Metric{
				arch:               info.arch,
				core:               fmt.Sprintf("%d", core_index),
				device:             info.device,
				kubernetesNodeName: t.nodeName,
				uuid:               info.uuid,
				cycleCount:         float64(0),
			}

			metric_map[core_index] = metric
			metricContainer = append(metricContainer, metric)
		}

		values, err := d.DevicePerformanceCounter()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		counters := values.PerformanceCounter()
		for _, counter := range counters {
			currentCycleCount := float64(counter.CycleCount())

			previousCycleCount := 0.0

			if value, ok := previousCycleCountMap[counter.Core()]; ok && previousCycleCount < currentCycleCount {
				previousCycleCount = value
			}

			metric_map[counter.Core()][cycleCount] = currentCycleCount - previousCycleCount

			previousCycleCountMap[counter.Core()] = currentCycleCount
		}
	}

	if err := t.postProcess(metricContainer); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (t *cycleCountCollector) postProcess(metrics MetricContainer) error {

	for _, metric := range metrics {
		if value, ok := metric[cycleCount]; ok {
			t.counterVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				core:               metric[core].(string),
				device:             metric[device].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
				uuid:               metric[uuid].(string),
			}).Add(value.(float64))
		}
	}

	return nil
}
