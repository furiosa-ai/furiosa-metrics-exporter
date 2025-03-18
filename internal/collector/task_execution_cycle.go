package collector

import (
	"errors"
	"fmt"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	taskExecutionCycle = "task_execution_cycle"
)

type taskExecutionCycleCollector struct {
	devices    []smi.Device
	counterVec *prometheus.CounterVec
	nodeName   string
}

var _ Collector = (*taskExecutionCycleCollector)(nil)

var (
	previousTaskExecutionCycleMap map[uint32]float64 = make(map[uint32]float64)
)

func NewTaskExecutionCycleCollector(devices []smi.Device, nodeName string) Collector {
	return &taskExecutionCycleCollector{
		devices:  devices,
		nodeName: nodeName,
	}
}

func (t *taskExecutionCycleCollector) Register() {
	t.counterVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "furiosa_npu_task_execution_cycle",
		Help: "The current task execution cycle of NPU device",
	},
		[]string{
			arch,
			device,
			core,
			kubernetesNodeName,
			uuid,
		})

}

func (t *taskExecutionCycleCollector) Collect() error {
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
				taskExecutionCycle: float64(0),
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
			currentTaskExecutionCycle := float64(counter.TaskExecutionCycle())

			previousTaskExecutionCycle := 0.0

			if value, ok := previousTaskExecutionCycleMap[counter.Core()]; ok && previousTaskExecutionCycle < currentTaskExecutionCycle {
				previousTaskExecutionCycle = value
			}

			metric_map[counter.Core()][taskExecutionCycle] = currentTaskExecutionCycle - previousTaskExecutionCycle

			previousTaskExecutionCycleMap[counter.Core()] = currentTaskExecutionCycle
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

func (t *taskExecutionCycleCollector) postProcess(metrics MetricContainer) error {

	for _, metric := range metrics {
		if value, ok := metric[taskExecutionCycle]; ok {
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
