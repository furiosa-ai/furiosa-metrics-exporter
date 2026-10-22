package collector

import (
	"errors"
	"sync"
	"time"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/prometheus/client_golang/prometheus"
)

type throttleEvent struct {
	timestamp      time.Time
	throttleReason api.ThrottleReason
	eventType      []string
}

type throttleReasonCollector struct {
	deviceSource  DeviceSource
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
	govClockCapLabel     = "gov_clock_cap"
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
	govClockCapLabel,
	otherReasonLabel,
}

var _ Collector = (*throttleReasonCollector)(nil)

func NewThrottleReasonCollector(deviceSource DeviceSource, metricFactory MetricFactory, kubeResMapper KubeResourcesMapper) Collector {
	newThrottleEvents := make(map[string][]throttleEvent)
	for _, d := range deviceSource.Devices() {
		newThrottleEvents[d.UUID()] = make([]throttleEvent, 0)
	}

	return &throttleReasonCollector{
		deviceSource:   deviceSource,
		metricFactory:  metricFactory,
		kubeResMapper:  kubeResMapper,
		interval:       1,
		windowSize:     300,
		throttleEvents: newThrottleEvents,
	}
}

func (t *throttleReasonCollector) GetDevices() []*api.Device {
	return t.deviceSource.Devices()
}

func (t *throttleReasonCollector) GetMetricFactory() MetricFactory {
	return t.metricFactory
}

func (t *throttleReasonCollector) Register() {
	go func() {
		tick := time.NewTicker(time.Second * time.Duration(t.interval))
		defer tick.Stop()

		for range tick.C {
			devices := t.GetDevices()

			t.throttleEventsLock.Lock()

			// Drop the events of devices that are no longer enumerated.
			current := make(map[string]struct{}, len(devices))
			for _, d := range devices {
				current[d.UUID()] = struct{}{}
			}

			for uuid := range t.throttleEvents {
				if _, ok := current[uuid]; !ok {
					delete(t.throttleEvents, uuid)
				}
			}

			for _, d := range devices {
				uuid := d.UUID()
				if _, ok := t.throttleEvents[uuid]; !ok {
					t.throttleEvents[uuid] = make([]throttleEvent, 0)
				}

				deviceThrottleReasons, err := d.ThrottleReason()
				if err != nil {
					continue
				}

				t.throttleEvents[uuid] = appendThrottleEvent(t.throttleEvents[uuid], foldThrottleReasons(deviceThrottleReasons))

				// Remove old events
				cutoff := time.Now().Add(-time.Duration(t.windowSize) * time.Second)
				events := t.throttleEvents[uuid]
				i := 0
				for ; i < len(events); i++ {
					if !events[i].timestamp.Before(cutoff) {
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

func (t *throttleReasonCollector) Collect(metrics map[*api.Device]Metric) error {
	devices := t.GetDevices()
	metricContainer := make(MetricContainer, 0, len(devices))

	errs := make([]error, 0)

	t.throttleEventsLock.RLock()

	for _, d := range devices {
		metric, exists := metrics[d]
		if !exists {
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

// foldThrottleReasons combines active throttle reasons into a single bitmask.
func foldThrottleReasons(reasons []api.ThrottleReason) api.ThrottleReason {
	var folded api.ThrottleReason
	for _, reason := range reasons {
		folded |= reason
	}

	return folded
}

func appendThrottleEvent(throttleEvents []throttleEvent, throttleReason api.ThrottleReason) []throttleEvent {
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

	if throttleReason&(api.ThrottleReasonIdle) != latestThrottleReason&(api.ThrottleReasonIdle) {
		newEventType = append(newEventType, idleLabel)
	}

	if throttleReason&(api.ThrottleReasonThermalSlowdown) != latestThrottleReason&(api.ThrottleReasonThermalSlowdown) {
		newEventType = append(newEventType, thermalSlowdownLabel)
	}

	if throttleReason&(api.ThrottleReasonAppPowerCap) != latestThrottleReason&(api.ThrottleReasonAppPowerCap) {
		newEventType = append(newEventType, appPowerCapLabel)
	}

	if throttleReason&(api.ThrottleReasonAppClockCap) != latestThrottleReason&(api.ThrottleReasonAppClockCap) {
		newEventType = append(newEventType, appClockCapLabel)
	}

	if throttleReason&(api.ThrottleReasonHwClockCap) != latestThrottleReason&(api.ThrottleReasonHwClockCap) {
		newEventType = append(newEventType, hwClockCapLabel)
	}

	if throttleReason&(api.ThrottleReasonHwBusLimit) != latestThrottleReason&(api.ThrottleReasonHwBusLimit) {
		newEventType = append(newEventType, hwBusLimitLabel)
	}

	if throttleReason&(api.ThrottleReasonHwPowerCap) != latestThrottleReason&(api.ThrottleReasonHwPowerCap) {
		newEventType = append(newEventType, hwPowerCapLabel)
	}

	if throttleReason&(api.ThrottleReasonGovClockCap) != latestThrottleReason&(api.ThrottleReasonGovClockCap) {
		newEventType = append(newEventType, govClockCapLabel)
	}

	if throttleReason&(api.ThrottleReasonOtherReason) != latestThrottleReason&(api.ThrottleReasonOtherReason) {
		newEventType = append(newEventType, otherReasonLabel)
	}

	newEvent := throttleEvent{
		timestamp:      time.Now(),
		throttleReason: throttleReason,
		eventType:      newEventType,
	}

	return append(throttleEvents, newEvent)
}
