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
	devices       []smi.Device
	metricFactory MetricFactory
	counterVec    *prometheus.CounterVec
}

var _ Collector = (*taskExecutionCycleCollector)(nil)

func NewTaskExecutionCycleCollector(devices []smi.Device, metricFactory MetricFactory) Collector {
	return &taskExecutionCycleCollector{
		devices:       devices,
		metricFactory: metricFactory,
	}
}

func (t *taskExecutionCycleCollector) Register() {
	t.counterVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "furiosa_npu_task_execution_cycle",
		Help: "The current task execution cycle of NPU device",
	}, defaultMetricLabels())
}

func (t *taskExecutionCycleCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		metric, err := t.metricFactory.NewDeviceWiseMetric(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		perfCounters, err := d.DevicePerformanceCounter()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		counters := perfCounters.PerformanceCounter()
		for _, counter := range counters {
			coreIndex := counter.Core()
			duplicated := deepCopyMetric(metric)
			duplicated[core] = strconv.Itoa(int(coreIndex))
			duplicated[taskExecutionCycle] = float64(counter.TaskExecutionCycle())
			metricContainer = append(metricContainer, duplicated)
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
	t.counterVec.Reset()
	transformed := TransformDeviceMetrics(metrics, true)

	for _, metric := range transformed {
		if value, ok := metric[taskExecutionCycle]; ok {
			t.counterVec.With(prometheus.Labels{
				arch:                metric[arch].(string),
				core:                metric[core].(string),
				device:              metric[device].(string),
				uuid:                metric[uuid].(string),
				bdf:                 metric[bdf].(string),
				firmwareVersion:     metric[firmwareVersion].(string),
				pertVersion:         metric[pertVersion].(string),
				driverVersion:       metric[driverVersion].(string),
				hostname:            metric[hostname].(string),
				kubernetesNamespace: metric[kubernetesNamespace].(string),
				kubernetesPod:       metric[kubernetesPod].(string),
				kubernetesContainer: metric[kubernetesContainer].(string),
			}).Add(value.(float64))
		}
	}

	return nil
}
