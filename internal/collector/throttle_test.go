package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func newFakeThrottleCollector() Collector {
	return &throttleReasonCollector{
		devices:        nil,
		metricFactory:  nil,
		kubeResMapper:  NewFakeKubeResourcesMapper(),
		interval:       1,
		windowSize:     300,
		throttleEvents: make(map[string][]throttleEvent),
	}
}

func TestThrottleCollector_PostProcessing(t *testing.T) {
	collector := newFakeThrottleCollector()
	collector.Register()

	tc := MetricContainer{}
	metric := newMetric()
	metric[arch] = "rngd"
	metric[core] = "0-7"
	metric[device] = "npu0"
	metric[uuid] = uuid
	metric[idleLabel] = float64(1)
	metric[thermalSlowdownLabel] = float64(1)
	metric[appPowerCapLabel] = float64(1)
	metric[appClockCapLabel] = float64(1)
	metric[hwClockCapLabel] = float64(1)
	metric[hwBusLimitLabel] = float64(1)
	metric[hwPowerCapLabel] = float64(1)
	metric[otherReasonLabel] = float64(1)

	tc = append(tc, metric)
	err := collector.postProcess(tc)
	assert.NoError(t, err)

	expected := `
# HELP furiosa_npu_throttling_events_count The throttling event count of NPU device
# TYPE furiosa_npu_throttling_events_count gauge
furiosa_npu_throttling_events_count{arch="rngd",core="0-7",device="npu0",label="app_clock_cap",uuid="uuid"} 1
furiosa_npu_throttling_events_count{arch="rngd",core="0-7",device="npu0",label="app_power_cap",uuid="uuid"} 1
furiosa_npu_throttling_events_count{arch="rngd",core="0-7",device="npu0",label="hw_bus_limit",uuid="uuid"} 1
furiosa_npu_throttling_events_count{arch="rngd",core="0-7",device="npu0",label="hw_clock_cap",uuid="uuid"} 1
furiosa_npu_throttling_events_count{arch="rngd",core="0-7",device="npu0",label="hw_power_cap",uuid="uuid"} 1
furiosa_npu_throttling_events_count{arch="rngd",core="0-7",device="npu0",label="idle",uuid="uuid"} 1
furiosa_npu_throttling_events_count{arch="rngd",core="0-7",device="npu0",label="other_reason",uuid="uuid"} 1
furiosa_npu_throttling_events_count{arch="rngd",core="0-7",device="npu0",label="thermal_slowdown",uuid="uuid"} 1
`
	err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "furiosa_npu_throttling_events_count")
	assert.NoError(t, err)
}

func TestThrottleCollector_Collect(t *testing.T) {
	//TODO: add testcases with device mock
}
