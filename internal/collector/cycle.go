package collector

import (
	"errors"
	"strconv"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	taskExecutionCycle = "task_execution_cycle"
	totalCycleCount    = "total_cycle_count"
)

type cycleCollector struct {
	devices       []smi.Device
	metricFactory MetricFactory
	kubeResMapper KubeResourcesMapper

	taskExecutionCycleCounterVec *prometheus.CounterVec
	totalCycleCountCounterVec    *prometheus.CounterVec
}

var _ Collector = (*cycleCollector)(nil)

func NewCycleCollector(devices []smi.Device, metricFactory MetricFactory, kubeResMapper KubeResourcesMapper) Collector {
	return &cycleCollector{
		devices:       devices,
		metricFactory: metricFactory,
		kubeResMapper: kubeResMapper,
	}
}

func (t *cycleCollector) Register() {
	taskExecutionCycleOpts := prometheus.CounterOpts{
		Name: "furiosa_npu_task_execution_cycle",
		Help: "The current task execution cycle of NPU device",
	}

	t.taskExecutionCycleCounterVec = prometheus.NewCounterVec(taskExecutionCycleOpts, defaultMetricLabels())

	prometheus.MustRegister(NewLabelFilterCollector(
		t.taskExecutionCycleCounterVec,
		prometheus.Opts(taskExecutionCycleOpts),
		prometheus.CounterValue,
	))

	totalCycleCountOpts := prometheus.CounterOpts{
		Name: "furiosa_npu_total_cycle_count",
		Help: "The current total cycle count of NPU device",
	}

	t.totalCycleCountCounterVec = prometheus.NewCounterVec(totalCycleCountOpts, defaultMetricLabels())
	prometheus.MustRegister(NewLabelFilterCollector(
		t.totalCycleCountCounterVec,
		prometheus.Opts(totalCycleCountOpts),
		prometheus.CounterValue,
	))
}

func (t *cycleCollector) Collect() error {
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
			duplicated[totalCycleCount] = float64(counter.CycleCount())

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

func (t *cycleCollector) postProcess(metrics MetricContainer) error {
	t.taskExecutionCycleCounterVec.Reset()
	t.totalCycleCountCounterVec.Reset()

	transformed := t.kubeResMapper.TransformDeviceMetrics(metrics, true)
	for _, metric := range transformed {
		if value, ok := metric[taskExecutionCycle]; ok {
			t.taskExecutionCycleCounterVec.With(prometheus.Labels{
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

		if value, ok := metric[totalCycleCount]; ok {
			t.totalCycleCountCounterVec.With(prometheus.Labels{
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
