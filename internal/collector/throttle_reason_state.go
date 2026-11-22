package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/prometheus/client_golang/prometheus"
)

// throttleReasonLabelByReason maps each known throttle reason to the label already used by
// furiosa_npu_throttling_events_count, so both metrics name a reason the same way.
var throttleReasonLabelByReason = map[api.ThrottleReason]string{
	api.ThrottleReasonIdle:            idleLabel,
	api.ThrottleReasonThermalSlowdown: thermalSlowdownLabel,
	api.ThrottleReasonAppPowerCap:     appPowerCapLabel,
	api.ThrottleReasonAppClockCap:     appClockCapLabel,
	api.ThrottleReasonHwClockCap:      hwClockCapLabel,
	api.ThrottleReasonHwBusLimit:      hwBusLimitLabel,
	api.ThrottleReasonHwPowerCap:      hwPowerCapLabel,
	api.ThrottleReasonGovClockCap:     govClockCapLabel,
	api.ThrottleReasonOtherReason:     otherReasonLabel,
}

type throttleReasonStateCollector struct {
	deviceSource  DeviceSource
	metricFactory MetricFactory
	gaugeVec      *prometheus.GaugeVec
	kubeResMapper KubeResourcesMapper
}

var _ Collector = (*throttleReasonStateCollector)(nil)

func NewThrottleReasonStateCollector(deviceSource DeviceSource, metricFactory MetricFactory, kubeResMapper KubeResourcesMapper) Collector {
	return &throttleReasonStateCollector{
		deviceSource:  deviceSource,
		metricFactory: metricFactory,
		kubeResMapper: kubeResMapper,
	}
}

func (t *throttleReasonStateCollector) GetDevices() []*api.Device {
	return t.deviceSource.Devices()
}

func (t *throttleReasonStateCollector) GetMetricFactory() MetricFactory {
	return t.metricFactory
}

func (t *throttleReasonStateCollector) Register() {
	opts := prometheus.GaugeOpts{
		Name: "furiosa_npu_throttle_reason",
		Help: "The current throttle state of NPU device",
	}

	t.gaugeVec = prometheus.NewGaugeVec(opts, append(defaultMetricLabels(), label))

	prometheus.MustRegister(NewLabelFilterCollector(
		t.gaugeVec,
		prometheus.Opts(opts),
		prometheus.GaugeValue,
	))
}

func (t *throttleReasonStateCollector) Collect(metrics map[*api.Device]Metric) error {
	devices := t.GetDevices()
	metricContainer := make(MetricContainer, 0, len(devices))

	errs := make([]error, 0)
	for _, d := range devices {
		metric, exists := metrics[d]
		if !exists {
			continue
		}

		reasons, err := d.ThrottleReason()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		setThrottleReasonState(metric, reasons)
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

func (t *throttleReasonStateCollector) postProcess(metrics MetricContainer) error {
	transformed := t.kubeResMapper.TransformDeviceMetrics(metrics, false)
	t.gaugeVec.Reset()

	for _, metric := range transformed {
		for _, throttleLabel := range throttleReasonLabels {
			if value, ok := metric[throttleLabel]; ok {
				t.gaugeVec.With(prometheus.Labels{
					arch:                metric[arch].(string),
					core:                metric[core].(string),
					device:              metric[device].(string),
					label:               throttleLabel,
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

// setThrottleReasonState stores one value per known throttle reason: 1 while the reason is
// active, 0 otherwise. Bits the library does not name are left out, as the events counter does.
func setThrottleReasonState(metric Metric, reasons []api.ThrottleReason) {
	for _, throttleLabel := range throttleReasonLabels {
		metric[throttleLabel] = float64(0)
	}

	for _, reason := range reasons {
		if throttleLabel, ok := throttleReasonLabelByReason[reason]; ok {
			metric[throttleLabel] = float64(1)
		}
	}
}
