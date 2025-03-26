package collector

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCycleCountCollector_PostProcessing(t *testing.T) {
	tests := []struct {
		description string
		source      MetricContainer
		expected    string
	}{
		{
			description: "random performance counter metrics",
			source: func() MetricContainer {
				tc := MetricContainer{}
				for i := 0; i < 8; i++ {
					tc = append(tc, Metric{
						arch:               "rngd",
						core:               strconv.Itoa(i),
						device:             "npu0",
						kubernetesNodeName: "node",
						uuid:               "uuid",
						cycleCount:         float64(5678),
					})
				}
				return tc
			}(),
			expected: `
# HELP furiosa_npu_cycle_count The current cycle count of NPU device
# TYPE furiosa_npu_cycle_count counter
furiosa_npu_cycle_count{arch="rngd",core="0",device="npu0",kubernetes_node_name="node",uuid="uuid"} 5678
furiosa_npu_cycle_count{arch="rngd",core="1",device="npu0",kubernetes_node_name="node",uuid="uuid"} 5678
furiosa_npu_cycle_count{arch="rngd",core="2",device="npu0",kubernetes_node_name="node",uuid="uuid"} 5678
furiosa_npu_cycle_count{arch="rngd",core="3",device="npu0",kubernetes_node_name="node",uuid="uuid"} 5678
furiosa_npu_cycle_count{arch="rngd",core="4",device="npu0",kubernetes_node_name="node",uuid="uuid"} 5678
furiosa_npu_cycle_count{arch="rngd",core="5",device="npu0",kubernetes_node_name="node",uuid="uuid"} 5678
furiosa_npu_cycle_count{arch="rngd",core="6",device="npu0",kubernetes_node_name="node",uuid="uuid"} 5678
furiosa_npu_cycle_count{arch="rngd",core="7",device="npu0",kubernetes_node_name="node",uuid="uuid"} 5678
`,
		},
	}

	cc := &cycleCountCollector{}
	cc.Register()
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			err := cc.postProcess(tc.source)
			assert.NoError(t, err)

			err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(head+tc.expected), "furiosa_npu_cycle_count")
			assert.NoError(t, err)
		})
	}
}

func TestCycleCountCollector_Collect(t *testing.T) {
	//TODO: add test cases with mock device data
}
