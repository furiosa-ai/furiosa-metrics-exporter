package collector

import (
	"errors"
	"strconv"

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
			uuid,
			kubernetesNode,
			kubernetesNamespace,
			kubernetesPod,
			kubernetesContainer,
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
			metric := newMetric()
			metric[arch] = info.arch
			metric[core] = strconv.Itoa(int(core_index))
			metric[device] = info.device
			metric[uuid] = info.uuid
			metric[cycleCount] = float64(0)

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
	transformed := TransformDeviceMetrics(metrics, true)

	for _, metric := range transformed {
		if value, ok := metric[taskExecutionCycle]; ok {
			t.counterVec.With(prometheus.Labels{
				arch:                metric[arch].(string),
				core:                metric[core].(string),
				device:              metric[device].(string),
				uuid:                metric[uuid].(string),
				kubernetesNode:      t.nodeName,
				kubernetesNamespace: metric[kubernetesNamespace].(string),
				kubernetesPod:       metric[kubernetesPod].(string),
				kubernetesContainer: metric[kubernetesContainer].(string),
			}).Add(value.(float64))
		}
	}

	return nil
}
