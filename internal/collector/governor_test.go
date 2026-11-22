package collector

import (
	"strings"
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

const governorHead = `
# HELP furiosa_npu_governor The governor profile of NPU device
# TYPE furiosa_npu_governor gauge
`

func newFakeGovernorCollector() Collector {
	return &governorCollector{
		deviceSource:  StaticDeviceSource(nil),
		metricFactory: nil,
		kubeResMapper: NewFakeKubeResourcesMapper(),
	}
}

func TestGovernorCollector_PostProcessing(t *testing.T) {
	tests := []struct {
		description string
		profile     api.GovernorProfile
		expected    string
	}{
		{
			description: "performance profile is active",
			profile:     api.GovernorProfilePerformance,
			expected: `
furiosa_npu_governor{arch="rngd",core="0-7",device="npu0",label="performance",uuid="uuid"} 1
furiosa_npu_governor{arch="rngd",core="0-7",device="npu0",label="powersave",uuid="uuid"} 0
`,
		},
		{
			description: "powersave profile is active",
			profile:     api.GovernorProfilePowerSave,
			expected: `
furiosa_npu_governor{arch="rngd",core="0-7",device="npu0",label="performance",uuid="uuid"} 0
furiosa_npu_governor{arch="rngd",core="0-7",device="npu0",label="powersave",uuid="uuid"} 1
`,
		},
	}

	collector := newFakeGovernorCollector()
	collector.Register()
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			metric := newMetric()
			metric[arch] = "rngd"
			metric[core] = "0-7"
			metric[device] = "npu0"
			metric[uuid] = "uuid"
			setGovernorProfile(metric, tc.profile)

			err := collector.postProcess(MetricContainer{metric})
			assert.NoError(t, err)

			err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(governorHead+tc.expected), "furiosa_npu_governor")
			assert.NoError(t, err)
		})
	}
}

func TestSetGovernorProfile_UnknownProfileExportsNoActiveSeries(t *testing.T) {
	metric := newMetric()
	setGovernorProfile(metric, api.GovernorProfile(99))

	assert.Equal(t, float64(0), metric[performanceLabel])
	assert.Equal(t, float64(0), metric[powersaveLabel])
}

func TestGovernorCollector_Collect(t *testing.T) {
	//TODO: add testcases with device mock
}
