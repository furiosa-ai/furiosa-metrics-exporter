package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	liveness = "liveness"
)

type livenessCollector struct {
	deviceSource  DeviceSource
	metricFactory MetricFactory
	gaugeVec      *prometheus.GaugeVec
	kubeResMapper KubeResourcesMapper
}

var _ Collector = (*livenessCollector)(nil)

func NewLivenessCollector(deviceSource DeviceSource, metricFactory MetricFactory, kubeResMapper KubeResourcesMapper) Collector {
	return &livenessCollector{
		deviceSource:  deviceSource,
		metricFactory: metricFactory,
		kubeResMapper: kubeResMapper,
	}
}

func (t *livenessCollector) GetDevices() []*api.Device {
	return t.deviceSource.Devices()
}

func (t *livenessCollector) GetMetricFactory() MetricFactory {
	return t.metricFactory
}

func (t *livenessCollector) Register() {
	opts := prometheus.GaugeOpts{
		Name: "furiosa_npu_alive",
		Help: "The liveness of NPU device",
	}

	t.gaugeVec = prometheus.NewGaugeVec(opts, defaultMetricLabels())

	prometheus.MustRegister(NewLabelFilterCollector(
		t.gaugeVec,
		prometheus.Opts(opts),
		prometheus.GaugeValue,
	))
}

func (t *livenessCollector) Collect(metrics map[*api.Device]Metric) error {
	devices := t.GetDevices()
	metricContainer := make(MetricContainer, 0, len(devices))

	errs := make([]error, 0)
	for _, d := range devices {
		metric, exists := metrics[d]
		if !exists {
			continue
		}

		metric[liveness] = d.Liveness()
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
	transformed := t.kubeResMapper.TransformDeviceMetrics(metrics, false)
	t.gaugeVec.Reset()

	for _, metric := range transformed {
		if value, ok := metric[liveness]; ok {
			var alive float64
			if value.(bool) {
				alive = 1
			} else {
				alive = 0
			}

			t.gaugeVec.With(prometheus.Labels{
				arch:                metric[arch].(string),
				core:                metric[core].(string),
				device:              metric[device].(string),
				uuid:                metric[uuid].(string),
				bdf:                 metric[bdf].(string),
				firmwareVersion:     metric[firmwareVersion].(string),
				driverVersion:       metric[driverVersion].(string),
				hostname:            metric[hostname].(string),
				kubernetesNamespace: metric[kubernetesNamespace].(string),
				kubernetesPod:       metric[kubernetesPod].(string),
				kubernetesContainer: metric[kubernetesContainer].(string),
			}).Set(alive)

		}
	}

	return nil
}
