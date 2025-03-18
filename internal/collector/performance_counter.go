package collector

import (
	"errors"
	"fmt"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	taskExecutionCycle = "take_execution_cycle"
	cycleCount         = "cycle_count"
)

type performanceCounterCollector struct {
	devices  []smi.Device
	gaugeVec *prometheus.GaugeVec
	nodeName string
}

var _ Collector = (*performanceCounterCollector)(nil)

func NewPerformanceCounterCollector(devices []smi.Device, nodeName string) Collector {
	return &performanceCounterCollector{
		devices:  devices,
		nodeName: nodeName,
	}
}

func (t *performanceCounterCollector) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_performance_counter",
		Help: "The current performance counter of NPU device",
	},
		[]string{
			arch,
			device,
			core,
			kubernetesNodeName,
			label,
			uuid,
		})
}

func (t *performanceCounterCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		info, err := getDeviceInfo(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		values, err := d.DevicePerformanceCounter()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		counters := values.PerformanceCounter()
		for _, counter := range counters {
			metric := Metric{
				arch:               info.arch,
				core:               fmt.Sprintf("%d", counter.Core()),
				device:             info.device,
				kubernetesNodeName: t.nodeName,
				uuid:               info.uuid,
				taskExecutionCycle: counter.TaskExecutionCycle(),
				cycleCount:         counter.CycleCount(),
			}
			metricContainer = append(metricContainer, metric)
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

func (t *performanceCounterCollector) postProcess(metrics MetricContainer) error {
	t.gaugeVec.Reset()

	for _, metric := range metrics {
		if value, ok := metric[taskExecutionCycle]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				core:               metric[core].(string),
				device:             metric[device].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
				label:              taskExecutionCycle,
				uuid:               metric[uuid].(string),
			}).Set(float64(value.(uint64)))
		}
		if value, ok := metric[cycleCount]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:               metric[arch].(string),
				core:               metric[core].(string),
				device:             metric[device].(string),
				kubernetesNodeName: metric[kubernetesNodeName].(string),
				label:              cycleCount,
				uuid:               metric[uuid].(string),
			}).Set(float64(value.(uint64)))
		}
	}

	return nil
}
