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
		devices:       nil,
		metricFactory: nil,
		kubeResMapper: NewFakeKubeResourcesMapper(),
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
	metric[idle] = float64(1)
	metric[thermalSlowdown] = float64(1)
	metric[appPowerCap] = float64(1)
	metric[appClockCap] = float64(1)
	metric[hwClockCap] = float64(1)
	metric[hwBusLimit] = float64(1)
	metric[hwPowerCap] = float64(1)
	metric[otherReason] = float64(1)

	tc = append(tc, metric)
	err := collector.postProcess(tc)
	assert.NoError(t, err)

	expected := `
# HELP furiosa_npu_throttling_events_count The throttling event count of NPU device
# TYPE furiosa_npu_throttling_events_count counter
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
