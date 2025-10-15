package collector

import (
	"errors"
	"time"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
)

type throttleEvent struct {
	timestamp      time.Time
	throttleReason smi.ThrottleReason
	eventCount     map[string]float64
}

type throttleReasonCollector struct {
	devices       []smi.Device
	metricFactory MetricFactory
	counterVec    *prometheus.CounterVec
	kubeResMapper KubeResourcesMapper

	interval   int
	windowSize int

	throttleEvents map[smi.Device][]throttleEvent
}

const (
	idleLabel            = "idle"
	thermalSlowdownLabel = "thermal_slowdown"
	appPowerCapLabel     = "app_power_cap"
	appClockCapLabel     = "app_clock_cap"
	hwClockCapLabel      = "hw_clock_cap"
	hwBusLimitLabel      = "hw_bus_limit"
	hwPowerCapLabel      = "hw_power_cap"
	otherReasonLabel     = "other_reason"
)

var throttleReasonLabels = []string{
	idleLabel,
	thermalSlowdownLabel,
	appPowerCapLabel,
	appClockCapLabel,
	hwClockCapLabel,
	hwBusLimitLabel,
	hwPowerCapLabel,
	otherReasonLabel,
}

var _ Collector = (*throttleReasonCollector)(nil)

func NewThrottleReasonCollector(devices []smi.Device, metricFactory MetricFactory, kubeResMapper KubeResourcesMapper) Collector {
	newThrottleEvents := make(map[smi.Device][]throttleEvent)
	for _, d := range devices {
		newThrottleEvents[d] = make([]throttleEvent, 0)
	}

	return &throttleReasonCollector{
		devices:        devices,
		metricFactory:  metricFactory,
		kubeResMapper:  kubeResMapper,
		interval:       1,
		windowSize:     300,
		throttleEvents: newThrottleEvents,
	}
}

func (t *throttleReasonCollector) Register() {
	go func() {
		tick := time.NewTicker(time.Second * time.Duration(t.interval))
		defer tick.Stop()

		for range tick.C {
			for _, d := range t.devices {
				deviceThrottleReason, err := d.ThrottleReason()
				if err != nil {
					continue
				}

				t.throttleEvents[d] = appendThrottleEvent(t.throttleEvents[d], deviceThrottleReason)

				// Remove old events
				cutoff := time.Now().Add(-time.Duration(t.windowSize) * time.Second)
				events := t.throttleEvents[d]
				i := 0
				for ; i < len(events); i++ {
					if events[i].timestamp.After(cutoff) {
						break
					}
				}
				t.throttleEvents[d] = events[i:]
			}
		}
	}()

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

		if len(t.throttleEvents[d]) == 0 {
			continue
		}

		latestThrottleEventCount := t.throttleEvents[d][len(t.throttleEvents[d])-1].eventCount

		for throttleLabel, count := range latestThrottleEventCount {
			metric[throttleLabel] = count
		}

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
	t.counterVec.Reset()

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

func appendThrottleEvent(throttleEvents []throttleEvent, throttleReason smi.ThrottleReason) []throttleEvent {
	if len(throttleEvents) == 0 {
		newEvent := throttleEvent{
			timestamp:      time.Now(),
			throttleReason: throttleReason,
			eventCount: map[string]float64{
				idleLabel:            0,
				thermalSlowdownLabel: 0,
				appPowerCapLabel:     0,
				appClockCapLabel:     0,
				hwClockCapLabel:      0,
				hwBusLimitLabel:      0,
				hwPowerCapLabel:      0,
				otherReasonLabel:     0,
			},
		}

		return []throttleEvent{
			newEvent,
		}
	}

	latestEvent := throttleEvents[len(throttleEvents)-1]
	latestThrottleReason := latestEvent.throttleReason

	if latestThrottleReason == throttleReason {
		return throttleEvents
	}

	newEventCount := make(map[string]float64)

	if throttleReason&(smi.ThrottleReasonIdle) != latestThrottleReason&(smi.ThrottleReasonIdle) {
		newEventCount[idleLabel] = latestEvent.eventCount[idleLabel] + 1
	} else {
		newEventCount[idleLabel] = latestEvent.eventCount[idleLabel]
	}

	if throttleReason&(smi.ThrottleReasonThermalSlowdown) != latestThrottleReason&(smi.ThrottleReasonThermalSlowdown) {
		newEventCount[thermalSlowdownLabel] = latestEvent.eventCount[thermalSlowdownLabel] + 1
	} else {
		newEventCount[thermalSlowdownLabel] = latestEvent.eventCount[thermalSlowdownLabel]
	}

	if throttleReason&(smi.ThrottleReasonAppPowerCap) != latestThrottleReason&(smi.ThrottleReasonAppPowerCap) {
		newEventCount[appPowerCapLabel] = latestEvent.eventCount[appPowerCapLabel] + 1
	} else {
		newEventCount[appPowerCapLabel] = latestEvent.eventCount[appPowerCapLabel]
	}

	if throttleReason&(smi.ThrottleReasonAppClockCap) != latestThrottleReason&(smi.ThrottleReasonAppClockCap) {
		newEventCount[appClockCapLabel] = latestEvent.eventCount[appClockCapLabel] + 1
	} else {
		newEventCount[appClockCapLabel] = latestEvent.eventCount[appClockCapLabel]
	}

	if throttleReason&(smi.ThrottleReasonHwClockCap) != latestThrottleReason&(smi.ThrottleReasonHwClockCap) {
		newEventCount[hwClockCapLabel] = latestEvent.eventCount[hwClockCapLabel] + 1
	} else {
		newEventCount[hwClockCapLabel] = latestEvent.eventCount[hwClockCapLabel]
	}

	if throttleReason&(smi.ThrottleReasonHwBusLimit) != latestThrottleReason&(smi.ThrottleReasonHwBusLimit) {
		newEventCount[hwBusLimitLabel] = latestEvent.eventCount[hwBusLimitLabel] + 1
	} else {
		newEventCount[hwBusLimitLabel] = latestEvent.eventCount[hwBusLimitLabel]
	}

	if throttleReason&(smi.ThrottleReasonHwPowerCap) != latestThrottleReason&(smi.ThrottleReasonHwPowerCap) {
		newEventCount[hwPowerCapLabel] = latestEvent.eventCount[hwPowerCapLabel] + 1
	} else {
		newEventCount[hwPowerCapLabel] = latestEvent.eventCount[hwPowerCapLabel]
	}

	if throttleReason&(smi.ThrottleReasonOtherReason) != latestThrottleReason&(smi.ThrottleReasonOtherReason) {
		newEventCount[otherReasonLabel] = latestEvent.eventCount[otherReasonLabel] + 1
	} else {
		newEventCount[otherReasonLabel] = latestEvent.eventCount[otherReasonLabel]
	}

	newEvent := throttleEvent{
		timestamp:      time.Now(),
		throttleReason: throttleReason,
		eventCount:     newEventCount,
	}

	return append(throttleEvents, newEvent)
}
