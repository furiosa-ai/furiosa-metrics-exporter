package collector

import (
	"errors"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	peUtilization = "peUtilization"
)

type coreUtilizationCollector struct {
	deviceNames   []string
	metricFactory MetricFactory
	gaugeVec      *prometheus.GaugeVec
}

var _ Collector = (*coreUtilizationCollector)(nil)

func NewCoreUtilizationCollector(deviceNames []string, metricFactory MetricFactory) Collector {
	return &coreUtilizationCollector{
		deviceNames:   deviceNames,
		metricFactory: metricFactory,
	}
}

func (t *coreUtilizationCollector) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_core_utilization",
		Help: "The current core utilization of NPU device",
	}, defaultMetricLabels())
}

func (t *coreUtilizationCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.deviceNames))

	errs := make([]error, 0)
	for _, deviceName := range t.deviceNames {
		metric, err := t.metricFactory.NewDeviceWiseMetric(deviceName)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		coreUtilization := DeviceSMICacheMap[deviceName].coreUtilization

		utilization := coreUtilization.PeUtilization()
		for _, pe := range utilization {
			duplicated := deepCopyMetric(metric)
			duplicated[core] = strconv.Itoa(int(pe.Core()))
			duplicated[peUtilization] = pe.PeUsagePercentage()
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

func (t *coreUtilizationCollector) postProcess(metrics MetricContainer) error {
	transformed := TransformDeviceMetrics(metrics, true)
	t.gaugeVec.Reset()

	for _, metric := range transformed {
		if value, ok := metric[peUtilization]; ok {
			t.gaugeVec.With(prometheus.Labels{
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
			}).Set(value.(float64))
		}
	}

	return nil
}
