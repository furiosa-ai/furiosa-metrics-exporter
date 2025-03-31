package collector

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	rms = "rms"
)

type powerCollector struct {
	deviceNames   []string
	metricFactory MetricFactory
	gaugeVec      *prometheus.GaugeVec
}

var _ Collector = (*powerCollector)(nil)

func NewPowerCollector(deviceNames []string, metricFactory MetricFactory) Collector {
	return &powerCollector{
		deviceNames:   deviceNames,
		metricFactory: metricFactory,
	}
}

func (t *powerCollector) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_hw_power",
		Help: "The current power of NPU device",
	}, append(defaultMetricLabels(), label))
}

func (t *powerCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.deviceNames))

	errs := make([]error, 0)
	for _, deviceName := range t.deviceNames {
		metric, err := t.metricFactory.NewDeviceWiseMetric(deviceName)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		power := DeviceSMICacheMap[deviceName].power

		metric[rms] = power
		metricContainer = append(metricContainer, metric)
	}

	if err := t.postProcess(metricContainer); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (t *powerCollector) postProcess(metrics MetricContainer) error {
	transformed := TransformDeviceMetrics(metrics, false)
	t.gaugeVec.Reset()

	for _, metric := range transformed {
		if value, ok := metric["rms"]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:                metric[arch].(string),
				core:                metric[core].(string),
				device:              metric[device].(string),
				label:               rms,
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
