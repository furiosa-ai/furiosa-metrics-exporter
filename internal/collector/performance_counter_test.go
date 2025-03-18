package collector

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPerformanceCounterCollector_PostProcessing(t *testing.T) {
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
						taskExecutionCycle: uint64(1234),
						cycleCount:         uint64(5678),
					})
				}
				return tc
			}(),
			expected: `
# HELP furiosa_npu_performance_counter The current performance counter of NPU device
# TYPE furiosa_npu_performance_counter gauge
furiosa_npu_performance_counter{arch="rngd",core="0",device="npu0",kubernetes_node_name="node",label="cycle_count",uuid="uuid"} 5678
furiosa_npu_performance_counter{arch="rngd",core="0",device="npu0",kubernetes_node_name="node",label="take_execution_cycle",uuid="uuid"} 1234
furiosa_npu_performance_counter{arch="rngd",core="1",device="npu0",kubernetes_node_name="node",label="cycle_count",uuid="uuid"} 5678
furiosa_npu_performance_counter{arch="rngd",core="1",device="npu0",kubernetes_node_name="node",label="take_execution_cycle",uuid="uuid"} 1234
furiosa_npu_performance_counter{arch="rngd",core="2",device="npu0",kubernetes_node_name="node",label="cycle_count",uuid="uuid"} 5678
furiosa_npu_performance_counter{arch="rngd",core="2",device="npu0",kubernetes_node_name="node",label="take_execution_cycle",uuid="uuid"} 1234
furiosa_npu_performance_counter{arch="rngd",core="3",device="npu0",kubernetes_node_name="node",label="cycle_count",uuid="uuid"} 5678
furiosa_npu_performance_counter{arch="rngd",core="3",device="npu0",kubernetes_node_name="node",label="take_execution_cycle",uuid="uuid"} 1234
furiosa_npu_performance_counter{arch="rngd",core="4",device="npu0",kubernetes_node_name="node",label="cycle_count",uuid="uuid"} 5678
furiosa_npu_performance_counter{arch="rngd",core="4",device="npu0",kubernetes_node_name="node",label="take_execution_cycle",uuid="uuid"} 1234
furiosa_npu_performance_counter{arch="rngd",core="5",device="npu0",kubernetes_node_name="node",label="cycle_count",uuid="uuid"} 5678
furiosa_npu_performance_counter{arch="rngd",core="5",device="npu0",kubernetes_node_name="node",label="take_execution_cycle",uuid="uuid"} 1234
furiosa_npu_performance_counter{arch="rngd",core="6",device="npu0",kubernetes_node_name="node",label="cycle_count",uuid="uuid"} 5678
furiosa_npu_performance_counter{arch="rngd",core="6",device="npu0",kubernetes_node_name="node",label="take_execution_cycle",uuid="uuid"} 1234
furiosa_npu_performance_counter{arch="rngd",core="7",device="npu0",kubernetes_node_name="node",label="cycle_count",uuid="uuid"} 5678
furiosa_npu_performance_counter{arch="rngd",core="7",device="npu0",kubernetes_node_name="node",label="take_execution_cycle",uuid="uuid"} 1234
`,
		},
	}

	cu := &performanceCounterCollector{}
	cu.Register()
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			err := cu.postProcess(tc.source)
			assert.NoError(t, err)

			err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(head+tc.expected), "furiosa_npu_performance_counter")
			assert.NoError(t, err)
		})
	}
}

func TestPerformanceCounterCollector_Collect(t *testing.T) {
	//TODO: add test cases with mock device data
}
