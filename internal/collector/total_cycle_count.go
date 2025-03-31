package collector

import (
	"errors"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	totalCycleCount = "total_cycle_count"
)

type totalCycleCountCollector struct {
	deviceNames   []string
	metricFactory MetricFactory
	counterVec    *prometheus.CounterVec
}

var _ Collector = (*totalCycleCountCollector)(nil)

func NewTotalCycleCountCollector(deviceNames []string, metricFactory MetricFactory) Collector {
	return &totalCycleCountCollector{
		deviceNames:   deviceNames,
		metricFactory: metricFactory,
	}
}

func (t *totalCycleCountCollector) Register() {
	t.counterVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "furiosa_npu_total_cycle_count",
		Help: "The current total cycle count of NPU device",
	}, defaultMetricLabels())
}

func (t *totalCycleCountCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.deviceNames))

	errs := make([]error, 0)
	for _, deviceName := range t.deviceNames {
		metric, err := t.metricFactory.NewDeviceWiseMetric(deviceName)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		perfCounters := DeviceSMICacheMap[deviceName].performanceCounter

		counters := perfCounters.PerformanceCounter()
		for _, counter := range counters {
			coreIndex := counter.Core()
			duplicated := deepCopyMetric(metric)
			duplicated[core] = strconv.Itoa(int(coreIndex))
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

func (t *totalCycleCountCollector) postProcess(metrics MetricContainer) error {
	t.counterVec.Reset()
	transformed := TransformDeviceMetrics(metrics, true)

	for _, metric := range transformed {
		if value, ok := metric[totalCycleCount]; ok {
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
