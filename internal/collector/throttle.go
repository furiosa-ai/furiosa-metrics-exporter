package collector

import (
	"errors"
	"sync"
	"time"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/prometheus/client_golang/prometheus"
)

type throttleEvent struct {
	timestamp      time.Time
	throttleReason smi.ThrottleReason
	eventType      []string
}

type throttleReasonCollector struct {
	devices       []smi.Device
	metricFactory MetricFactory
	gaugeVec      *prometheus.GaugeVec
	kubeResMapper KubeResourcesMapper

	interval   int
	windowSize int

	throttleEventsLock sync.RWMutex
	throttleEvents     map[string][]throttleEvent
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
	newThrottleEvents := make(map[string][]throttleEvent)
	for _, d := range devices {
		deviceInfo, err := d.DeviceInfo()
		if err != nil {
			continue
		}

		uuid := deviceInfo.UUID()

		newThrottleEvents[uuid] = make([]throttleEvent, 0)
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
			t.throttleEventsLock.Lock()

			for _, d := range t.devices {
				deviceThrottleReason, err := d.ThrottleReason()
				if err != nil {
					continue
				}

				deviceInfo, err := d.DeviceInfo()
				if err != nil {
					continue
				}

				uuid := deviceInfo.UUID()

				t.throttleEvents[uuid] = appendThrottleEvent(t.throttleEvents[uuid], deviceThrottleReason)

				// Remove old events
				cutoff := time.Now().Add(-time.Duration(t.windowSize) * time.Second)
				events := t.throttleEvents[uuid]
				i := 0
				for ; i < len(events); i++ {
					if events[i].timestamp.After(cutoff) {
						break
					}
				}
				t.throttleEvents[uuid] = events[i:]
			}

			t.throttleEventsLock.Unlock()
		}
	}()

	opts := prometheus.GaugeOpts{
		Name: "furiosa_npu_throttling_events_count",
		Help: "The throttling event count of NPU device",
	}

	t.gaugeVec = prometheus.NewGaugeVec(opts, append(defaultMetricLabels(), label))

	prometheus.MustRegister(NewLabelFilterCollector(
		t.gaugeVec,
		prometheus.Opts(opts),
		prometheus.GaugeValue,
	))
}

func (t *throttleReasonCollector) Collect() error {
	metricContainer := make(MetricContainer, 0, len(t.devices))

	errs := make([]error, 0)

	t.throttleEventsLock.RLock()

	for _, d := range t.devices {
		metric, err := t.metricFactory.NewDeviceWiseMetric(d)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		uuid := metric[uuid].(string)

		if _, ok := t.throttleEvents[uuid]; !ok {
			continue
		}

		for _, throttleLabel := range throttleReasonLabels {
			metric[throttleLabel] = float64(0)
		}

		for _, throttleEvent := range t.throttleEvents[uuid] {
			for _, throttleLabel := range throttleEvent.eventType {
				metric[throttleLabel] = metric[throttleLabel].(float64) + 1
			}
		}

		metricContainer = append(metricContainer, metric)
	}

	t.throttleEventsLock.RUnlock()

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

func appendThrottleEvent(throttleEvents []throttleEvent, throttleReason smi.ThrottleReason) []throttleEvent {
	if len(throttleEvents) == 0 {
		newEvent := throttleEvent{
			timestamp:      time.Now(),
			throttleReason: throttleReason,
			eventType:      []string{},
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

	newEventType := []string{}

	if throttleReason&(smi.ThrottleReasonIdle) != latestThrottleReason&(smi.ThrottleReasonIdle) {
		newEventType = append(newEventType, idleLabel)
	}

	if throttleReason&(smi.ThrottleReasonThermalSlowdown) != latestThrottleReason&(smi.ThrottleReasonThermalSlowdown) {
		newEventType = append(newEventType, thermalSlowdownLabel)
	}

	if throttleReason&(smi.ThrottleReasonAppPowerCap) != latestThrottleReason&(smi.ThrottleReasonAppPowerCap) {
		newEventType = append(newEventType, appPowerCapLabel)
	}

	if throttleReason&(smi.ThrottleReasonAppClockCap) != latestThrottleReason&(smi.ThrottleReasonAppClockCap) {
		newEventType = append(newEventType, appClockCapLabel)
	}

	if throttleReason&(smi.ThrottleReasonHwClockCap) != latestThrottleReason&(smi.ThrottleReasonHwClockCap) {
		newEventType = append(newEventType, hwClockCapLabel)
	}

	if throttleReason&(smi.ThrottleReasonHwBusLimit) != latestThrottleReason&(smi.ThrottleReasonHwBusLimit) {
		newEventType = append(newEventType, hwBusLimitLabel)
	}

	if throttleReason&(smi.ThrottleReasonHwPowerCap) != latestThrottleReason&(smi.ThrottleReasonHwPowerCap) {
		newEventType = append(newEventType, hwPowerCapLabel)
	}

	if throttleReason&(smi.ThrottleReasonOtherReason) != latestThrottleReason&(smi.ThrottleReasonOtherReason) {
		newEventType = append(newEventType, otherReasonLabel)
	}

	newEvent := throttleEvent{
		timestamp:      time.Now(),
		throttleReason: throttleReason,
		eventType:      newEventType,
	}

	return append(throttleEvents, newEvent)
}
