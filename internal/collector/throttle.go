package collector

import (
	"errors"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
)

type throttleReasonCollector struct {
	devices       []smi.Device
	metricFactory MetricFactory
	counterVec    *prometheus.CounterVec
	kubeResMapper KubeResourcesMapper
}

const (
	idle            = "idle"
	thermalSlowdown = "thermal_slowdown"
	appPowerCap     = "app_power_cap"
	appClockCap     = "app_clock_cap"
	hwClockCap      = "hw_clock_cap"
	hwBusLimit      = "hw_bus_limit"
	hwPowerCap      = "hw_power_cap"
	otherReason     = "other_reason"
)

var throttleReasonLabels = []string{
	idle,
	thermalSlowdown,
	appPowerCap,
	appClockCap,
	hwClockCap,
	hwBusLimit,
	hwPowerCap,
	otherReason,
}

var throttleReasonCache smi.ThrottleReason

var _ Collector = (*throttleReasonCollector)(nil)

func NewThrottleReasonCollector(devices []smi.Device, metricFactory MetricFactory, kubeResMapper KubeResourcesMapper) Collector {
	return &throttleReasonCollector{
		devices:       devices,
		metricFactory: metricFactory,
		kubeResMapper: kubeResMapper,
	}
}

func (t *throttleReasonCollector) Register() {
	opts := prometheus.CounterOpts{
		Name: "furiosa_npu_throttling_events_count",
		Help: "The throttling event count of NPU device",
	}

	t.counterVec = prometheus.NewCounterVec(opts, append(defaultMetricLabels(), label))

	prometheus.MustRegister(NewLabelFilterCollector(
		t.counterVec,
		prometheus.Opts(opts),
		prometheus.CounterValue,
	))
}

func (t *throttleReasonCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)
	for _, d := range t.devices {
		metric, err := t.metricFactory.NewDeviceWiseMetric(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		deviceThrottleReason, err := d.ThrottleReason()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		metric = appendThrottleReasonLabels(metric, deviceThrottleReason)

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

func (t *throttleReasonCollector) postProcess(metrics MetricContainer) error {
	transformed := t.kubeResMapper.TransformDeviceMetrics(metrics, false)

	for _, metric := range transformed {
		for _, throttleLabel := range throttleReasonLabels {
			if value, ok := metric[throttleLabel]; ok {
				t.counterVec.With(prometheus.Labels{
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
				}).Add(value.(float64))
			}
		}
	}

	return nil
}

func appendThrottleReasonLabels(metric Metric, throttleReason smi.ThrottleReason) Metric {
	if throttleReason&(smi.ThrottleReasonIdle) != throttleReasonCache&(smi.ThrottleReasonIdle) {
		metric[idle] = 1.0
	} else {
		metric[idle] = 0.0
	}

	if throttleReason&(smi.ThrottleReasonThermalSlowdown) != throttleReasonCache&(smi.ThrottleReasonThermalSlowdown) {
		metric[thermalSlowdown] = 1.0
	} else {
		metric[thermalSlowdown] = 0.0
	}

	if throttleReason&(smi.ThrottleReasonAppPowerCap) != throttleReasonCache&(smi.ThrottleReasonAppPowerCap) {
		metric[appPowerCap] = 1.0
	} else {
		metric[appPowerCap] = 0.0
	}

	if throttleReason&(smi.ThrottleReasonAppClockCap) != throttleReasonCache&(smi.ThrottleReasonAppClockCap) {
		metric[appClockCap] = 1.0
	} else {
		metric[appClockCap] = 0.0
	}

	if throttleReason&(smi.ThrottleReasonHwClockCap) != throttleReasonCache&(smi.ThrottleReasonHwClockCap) {
		metric[hwClockCap] = 1.0
	} else {
		metric[hwClockCap] = 0.0
	}

	if throttleReason&(smi.ThrottleReasonHwBusLimit) != throttleReasonCache&(smi.ThrottleReasonHwBusLimit) {
		metric[hwBusLimit] = 1.0
	} else {
		metric[hwBusLimit] = 0.0
	}

	if throttleReason&(smi.ThrottleReasonHwPowerCap) != throttleReasonCache&(smi.ThrottleReasonHwPowerCap) {
		metric[hwPowerCap] = 1.0
	} else {
		metric[hwPowerCap] = 0.0
	}

	if throttleReason&(smi.ThrottleReasonOtherReason) != throttleReasonCache&(smi.ThrottleReasonOtherReason) {
		metric[otherReason] = 1.0
	} else {
		metric[otherReason] = 0.0
	}

	return metric
}
