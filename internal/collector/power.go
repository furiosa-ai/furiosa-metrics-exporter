package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	rms = "rms"
)

type powerCollector struct {
	devices         []smi.Device
	metricFactory   MetricFactory
	gaugeVec        *prometheus.GaugeVec
	gaugeVecWithPod *prometheus.GaugeVec
}

var _ Collector = (*powerCollector)(nil)

func NewPowerCollector(devices []smi.Device, metricFactory MetricFactory) Collector {
	return &powerCollector{
		devices:       devices,
		metricFactory: metricFactory,
	}
}

func (t *powerCollector) Register(registryWithPod *prometheus.Registry) {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_hw_power",
		Help: "The current power of NPU device",
	}, append(defaultMetricLabels(), label))

	t.gaugeVecWithPod = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_hw_power",
		Help: "The current power of NPU device",
	}, append(defaultMetricLabelsWithPod(), label))
	registryWithPod.MustRegister(t.gaugeVecWithPod)
}

func (t *powerCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		metric, err := t.metricFactory.NewDeviceWiseMetric(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		power, err := d.PowerConsumption()
		if err != nil {
			errs = append(errs, err)
			continue
		}

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
	t.gaugeVecWithPod.Reset()

	for _, metric := range transformed {
		if value, ok := metric["rms"]; ok {
			if podname, ok := metric[kubernetesPod].(string); !ok || (ok && podname == "") {
				t.gaugeVec.With(prometheus.Labels{
					arch:            metric[arch].(string),
					core:            metric[core].(string),
					device:          metric[device].(string),
					label:           rms,
					uuid:            metric[uuid].(string),
					bdf:             metric[bdf].(string),
					firmwareVersion: metric[firmwareVersion].(string),
					pertVersion:     metric[pertVersion].(string),
					driverVersion:   metric[driverVersion].(string),
					hostname:        metric[hostname].(string),
				}).Set(value.(float64))
			} else {
				t.gaugeVecWithPod.With(prometheus.Labels{
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
	}

	return nil
}
