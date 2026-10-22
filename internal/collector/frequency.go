package collector

import (
	"errors"
	"strconv"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	peFrequency = "coreFrequency"
)

type coreFrequencyCollector struct {
	deviceSource  DeviceSource
	metricFactory MetricFactory
	gaugeVec      *prometheus.GaugeVec
	kubeResMapper KubeResourcesMapper
}

var _ Collector = (*coreFrequencyCollector)(nil)

func NewCoreFrequencyCollector(deviceSource DeviceSource, metricFactory MetricFactory, kubeResMapper KubeResourcesMapper) Collector {
	return &coreFrequencyCollector{
		deviceSource:  deviceSource,
		metricFactory: metricFactory,
		kubeResMapper: kubeResMapper,
	}
}

func (t *coreFrequencyCollector) GetDevices() []*api.Device {
	return t.deviceSource.Devices()
}

func (t *coreFrequencyCollector) GetMetricFactory() MetricFactory {
	return t.metricFactory
}

func (t *coreFrequencyCollector) Register() {
	opts := prometheus.GaugeOpts{
		Name: "furiosa_npu_core_frequency",
		Help: "The current core frequency of NPU device (MHz)",
	}

	t.gaugeVec = prometheus.NewGaugeVec(opts, defaultMetricLabels())

	prometheus.MustRegister(NewLabelFilterCollector(
		t.gaugeVec,
		prometheus.Opts(opts),
		prometheus.GaugeValue,
	))
}

func (t *coreFrequencyCollector) Collect(metrics map[*api.Device]Metric) error {
	devices := t.GetDevices()
	metricContainer := make(MetricContainer, 0, len(devices))

	errs := make([]error, 0)
	for _, d := range devices {
		metric, exists := metrics[d]
		if !exists {
			continue
		}

		coreFrequency, err := d.CoreFrequency()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		for coreIndex, frequency := range coreFrequency {
			duplicated := deepCopyMetric(metric)
			duplicated[core] = strconv.Itoa(int(coreIndex))
			duplicated[peFrequency] = frequency
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

func (t *coreFrequencyCollector) postProcess(metrics MetricContainer) error {
	transformed := t.kubeResMapper.TransformDeviceMetrics(metrics, true)
	t.gaugeVec.Reset()

	for _, metric := range transformed {
		if value, ok := metric[peFrequency]; ok {
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
			}).Set(float64(value.(uint32)))
		}
	}

	return nil
}
