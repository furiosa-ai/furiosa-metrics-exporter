package collector

import (
	"errors"
	"strconv"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	peUtilization = "peUtilization"
)

type coreUtilizationCollector struct {
	deviceSource  DeviceSource
	sampler       CoreUtilizationSampler
	metricFactory MetricFactory
	gaugeVec      *prometheus.GaugeVec
	kubeResMapper KubeResourcesMapper
}

var _ Collector = (*coreUtilizationCollector)(nil)

func NewCoreUtilizationCollector(deviceSource DeviceSource, sampler CoreUtilizationSampler, metricFactory MetricFactory, kubeResMapper KubeResourcesMapper) Collector {
	return &coreUtilizationCollector{
		deviceSource:  deviceSource,
		sampler:       sampler,
		metricFactory: metricFactory,
		kubeResMapper: kubeResMapper,
	}
}

func (t *coreUtilizationCollector) GetDevices() []*api.Device {
	return t.deviceSource.Devices()
}

func (t *coreUtilizationCollector) GetMetricFactory() MetricFactory {
	return t.metricFactory
}

func (t *coreUtilizationCollector) Register() {
	opts := prometheus.GaugeOpts{
		Name: "furiosa_npu_core_utilization",
		Help: "The current core utilization of NPU device",
	}

	t.gaugeVec = prometheus.NewGaugeVec(opts, defaultMetricLabels())

	prometheus.MustRegister(NewLabelFilterCollector(
		t.gaugeVec,
		prometheus.Opts(opts),
		prometheus.GaugeValue,
	))
}

func (t *coreUtilizationCollector) Collect(metrics map[*api.Device]Metric) error {
	devices := t.GetDevices()
	metricContainer := make(MetricContainer, 0, len(devices))

	errs := make([]error, 0)
	for _, d := range devices {
		metric, exists := metrics[d]
		if !exists {
			continue
		}

		coreUtilizationSlice, err := t.sampler.GetCoreUtilization(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		for _, coreUtilization := range coreUtilizationSlice {
			duplicated := deepCopyMetric(metric)
			duplicated[core] = strconv.Itoa(int(coreUtilization.Core))
			duplicated[peUtilization] = coreUtilization.PeUsagePercentage
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
	transformed := t.kubeResMapper.TransformDeviceMetrics(metrics, true)
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
