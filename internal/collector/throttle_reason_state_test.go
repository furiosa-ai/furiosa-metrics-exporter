package collector

import (
	"strings"
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

const throttleReasonStateHead = `
# HELP furiosa_npu_throttle_reason The current throttle state of NPU device
# TYPE furiosa_npu_throttle_reason gauge
`

func newFakeThrottleReasonStateCollector() Collector {
	return &throttleReasonStateCollector{
		deviceSource:  StaticDeviceSource(nil),
		metricFactory: nil,
		kubeResMapper: NewFakeKubeResourcesMapper(),
	}
}

func throttleReasonStateSeries(active map[string]bool) string {
	var b strings.Builder
	for _, throttleLabel := range []string{
		appClockCapLabel, appPowerCapLabel, govClockCapLabel, hwBusLimitLabel, hwClockCapLabel,
		hwPowerCapLabel, idleLabel, otherReasonLabel, thermalSlowdownLabel,
	} {
		value := "0"
		if active[throttleLabel] {
			value = "1"
		}
		b.WriteString(`furiosa_npu_throttle_reason{arch="rngd",core="0-7",device="npu0",label="` + throttleLabel + `",uuid="uuid"} ` + value + "\n")
	}
	return b.String()
}

func TestThrottleReasonStateCollector_PostProcessing(t *testing.T) {
	tests := []struct {
		description string
		reasons     []api.ThrottleReason
		active      map[string]bool
	}{
		{
			description: "no reason is active",
			reasons:     nil,
			active:      map[string]bool{},
		},
		{
			description: "one reason is active",
			reasons:     []api.ThrottleReason{api.ThrottleReasonThermalSlowdown},
			active:      map[string]bool{thermalSlowdownLabel: true},
		},
		{
			description: "two reasons are active at the same time",
			reasons:     []api.ThrottleReason{api.ThrottleReasonIdle, api.ThrottleReasonGovClockCap},
			active:      map[string]bool{idleLabel: true, govClockCapLabel: true},
		},
		{
			description: "an unrecognized bit is ignored",
			reasons:     []api.ThrottleReason{api.ThrottleReason(1 << 20), api.ThrottleReasonHwPowerCap},
			active:      map[string]bool{hwPowerCapLabel: true},
		},
	}

	collector := newFakeThrottleReasonStateCollector()
	collector.Register()
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			metric := newMetric()
			metric[arch] = "rngd"
			metric[core] = "0-7"
			metric[device] = "npu0"
			metric[uuid] = "uuid"
			setThrottleReasonState(metric, tc.reasons)

			err := collector.postProcess(MetricContainer{metric})
			assert.NoError(t, err)

			expected := throttleReasonStateHead + throttleReasonStateSeries(tc.active)
			err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "furiosa_npu_throttle_reason")
			assert.NoError(t, err)
		})
	}
}

func TestThrottleReasonLabelByReason_CoversEveryLabel(t *testing.T) {
	seen := make(map[string]bool, len(throttleReasonLabelByReason))
	for _, throttleLabel := range throttleReasonLabelByReason {
		seen[throttleLabel] = true
	}

	for _, throttleLabel := range throttleReasonLabels {
		assert.True(t, seen[throttleLabel], "label %q has no api.ThrottleReason mapping", throttleLabel)
	}
	assert.Len(t, throttleReasonLabelByReason, len(throttleReasonLabels))
}

func TestThrottleReasonStateCollector_Collect(t *testing.T) {
	//TODO: add testcases with device mock
}
