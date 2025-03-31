package collector

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type temperatureCollector struct {
	deviceNames   []string
	metricFactory MetricFactory
	gaugeVec      *prometheus.GaugeVec
}

const (
	ambient = "ambient"
	peak    = "peak"
)

var _ Collector = (*temperatureCollector)(nil)

func NewTemperatureCollector(deviceNames []string, metricFactory MetricFactory) Collector {
	return &temperatureCollector{
		deviceNames:   deviceNames,
		metricFactory: metricFactory,
	}
}

func (t *temperatureCollector) Register() {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_hw_temperature",
		Help: "The current temperature of NPU device",
	}, append(defaultMetricLabels(), label))
}

func (t *temperatureCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.deviceNames))

	errs := make([]error, 0)
	for _, deviceName := range t.deviceNames {
		metric, err := t.metricFactory.NewDeviceWiseMetric(deviceName)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		deviceTemperature := DeviceSMICacheMap[deviceName].temperature

		metric[ambient] = deviceTemperature.Ambient()
		metric[peak] = deviceTemperature.SocPeak()
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

func (t *temperatureCollector) postProcess(metrics MetricContainer) error {
	transformed := TransformDeviceMetrics(metrics, false)
	t.gaugeVec.Reset()

	for _, metric := range transformed {
		if value, ok := metric[ambient]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:                metric[arch].(string),
				core:                metric[core].(string),
				device:              metric[device].(string),
				label:               ambient,
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

		if value, ok := metric[peak]; ok {
			t.gaugeVec.With(prometheus.Labels{
				arch:                metric[arch].(string),
				core:                metric[core].(string),
				device:              metric[device].(string),
				label:               peak,
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
