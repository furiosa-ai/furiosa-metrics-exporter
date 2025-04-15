package collector

import (
	"errors"
	"strconv"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	peFrequency = "coreFrequency"
)

type coreFrequencyCollector struct {
	devices       []smi.Device
	metricFactory MetricFactory
	gaugeVec      *prometheus.GaugeVec
}

var _ Collector = (*coreFrequencyCollector)(nil)

func NewCoreFrequencyCollector(devices []smi.Device, metricFactory MetricFactory) Collector {
	return &coreFrequencyCollector{
		devices:       devices,
		metricFactory: metricFactory,
	}
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

func (t *coreFrequencyCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		metric, err := t.metricFactory.NewDeviceWiseMetric(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		coreFrequency, err := d.CoreFrequency()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		frequency := coreFrequency.PeFrequency()
		for _, pe := range frequency {
			duplicated := deepCopyMetric(metric)
			duplicated[core] = strconv.Itoa(int(pe.Core()))
			duplicated[peFrequency] = pe.Frequency()
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
	transformed := TransformDeviceMetrics(metrics, true)
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
				pertVersion:         metric[pertVersion].(string),
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
