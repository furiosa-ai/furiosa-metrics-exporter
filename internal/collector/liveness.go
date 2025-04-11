package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	liveness = "liveness"
)

type livenessCollector struct {
	devices         []smi.Device
	metricFactory   MetricFactory
	gaugeVec        *prometheus.GaugeVec
	gaugeVecWithPod *prometheus.GaugeVec
}

var _ Collector = (*livenessCollector)(nil)

func NewLivenessCollector(devices []smi.Device, metricFactory MetricFactory) Collector {
	return &livenessCollector{
		devices:       devices,
		metricFactory: metricFactory,
	}
}

func (t *livenessCollector) Register(registryWithPod *prometheus.Registry) {
	t.gaugeVec = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_alive",
		Help: "The liveness of NPU device",
	}, defaultMetricLabels())

	t.gaugeVecWithPod = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "furiosa_npu_alive",
		Help: "The liveness of NPU device",
	}, defaultMetricLabelsWithPod())
	registryWithPod.MustRegister(t.gaugeVecWithPod)
}

func (t *livenessCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		metric, err := t.metricFactory.NewDeviceWiseMetric(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		value, err := d.Liveness()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		metric[liveness] = value
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

func (t *livenessCollector) postProcess(metrics MetricContainer) error {
	transformed := TransformDeviceMetrics(metrics, false)
	t.gaugeVec.Reset()
	t.gaugeVecWithPod.Reset()

	for _, metric := range transformed {
		if value, ok := metric[liveness]; ok {
			var alive float64
			if value.(bool) {
				alive = 1
			} else {
				alive = 0
			}
			if podname, ok := metric[kubernetesPod].(string); !ok || (ok && podname == "") {
				t.gaugeVec.With(prometheus.Labels{
					arch:            metric[arch].(string),
					core:            metric[core].(string),
					device:          metric[device].(string),
					uuid:            metric[uuid].(string),
					bdf:             metric[bdf].(string),
					firmwareVersion: metric[firmwareVersion].(string),
					pertVersion:     metric[pertVersion].(string),
					driverVersion:   metric[driverVersion].(string),
					hostname:        metric[hostname].(string),
				}).Set(alive)
			} else {
				t.gaugeVecWithPod.With(prometheus.Labels{
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
				}).Set(alive)
			}

		}
	}

	return nil
}
