package collector

import (
	"errors"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
)

type temperatureCollector struct {
	devices       []smi.Device
	metricFactory MetricFactory
	gaugeVec      *prometheus.GaugeVec
	kubeResMapper KubeResourcesMapper
}

const (
	ambient = "ambient"
	peak    = "peak"
)

var _ Collector = (*temperatureCollector)(nil)

func NewTemperatureCollector(devices []smi.Device, metricFactory MetricFactory, kubeResMapper KubeResourcesMapper) Collector {
	return &temperatureCollector{
		devices:       devices,
		metricFactory: metricFactory,
		kubeResMapper: kubeResMapper,
	}
}

func (t *temperatureCollector) Register() {
	opts := prometheus.GaugeOpts{
		Name: "furiosa_npu_hw_temperature",
		Help: "The current temperature of NPU device",
	}

	t.gaugeVec = prometheus.NewGaugeVec(opts, append(defaultMetricLabels(), label))

	prometheus.MustRegister(NewLabelFilterCollector(
		t.gaugeVec,
		prometheus.Opts(opts),
		prometheus.GaugeValue,
	))
}

func (t *temperatureCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		metric, err := t.metricFactory.NewDeviceWiseMetric(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		deviceTemperature, err := d.DeviceTemperature()
		if err != nil {
			errs = append(errs, err)
			continue
		}

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
	transformed := t.kubeResMapper.TransformDeviceMetrics(metrics, false)
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
