package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	performanceLabel = "performance"
	powersaveLabel   = "powersave"
)

var governorProfileLabels = []string{
	performanceLabel,
	powersaveLabel,
}

type governorCollector struct {
	deviceSource  DeviceSource
	metricFactory MetricFactory
	gaugeVec      *prometheus.GaugeVec
	kubeResMapper KubeResourcesMapper
}

var _ Collector = (*governorCollector)(nil)

func NewGovernorCollector(deviceSource DeviceSource, metricFactory MetricFactory, kubeResMapper KubeResourcesMapper) Collector {
	return &governorCollector{
		deviceSource:  deviceSource,
		metricFactory: metricFactory,
		kubeResMapper: kubeResMapper,
	}
}

func (t *governorCollector) GetDevices() []*api.Device {
	return t.deviceSource.Devices()
}

func (t *governorCollector) GetMetricFactory() MetricFactory {
	return t.metricFactory
}

func (t *governorCollector) Register() {
	opts := prometheus.GaugeOpts{
		Name: "furiosa_npu_governor",
		Help: "The governor profile of NPU device",
	}

	t.gaugeVec = prometheus.NewGaugeVec(opts, append(defaultMetricLabels(), label))

	prometheus.MustRegister(NewLabelFilterCollector(
		t.gaugeVec,
		prometheus.Opts(opts),
		prometheus.GaugeValue,
	))
}

func (t *governorCollector) Collect(metrics map[*api.Device]Metric) error {
	devices := t.GetDevices()
	metricContainer := make(MetricContainer, 0, len(devices))

	errs := make([]error, 0)
	for _, d := range devices {
		metric, exists := metrics[d]
		if !exists {
			continue
		}

		profile, err := d.GovernorProfile()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		setGovernorProfile(metric, profile)
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

func (t *governorCollector) postProcess(metrics MetricContainer) error {
	transformed := t.kubeResMapper.TransformDeviceMetrics(metrics, false)
	t.gaugeVec.Reset()

	for _, metric := range transformed {
		for _, profileLabel := range governorProfileLabels {
			if value, ok := metric[profileLabel]; ok {
				t.gaugeVec.With(prometheus.Labels{
					arch:                metric[arch].(string),
					core:                metric[core].(string),
					device:              metric[device].(string),
					label:               profileLabel,
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
	}

	return nil
}

// setGovernorProfile stores one value per governor profile: 1 for the active profile, 0 for the others.
func setGovernorProfile(metric Metric, profile api.GovernorProfile) {
	for _, profileLabel := range governorProfileLabels {
		metric[profileLabel] = float64(0)
	}

	switch profile {
	case api.GovernorProfilePerformance:
		metric[performanceLabel] = float64(1)
	case api.GovernorProfilePowerSave:
		metric[powersaveLabel] = float64(1)
	}
}
