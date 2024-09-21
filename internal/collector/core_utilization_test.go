package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"strconv"
	"strings"
	"testing"
)

func TestCoreUtilizationCollector_PostProcessing(t *testing.T) {
	tests := []struct {
		description string
		source      MetricContainer
		expected    string
	}{
		{
			description: "random core utilization metrics",
			source: func() MetricContainer {
				tc := MetricContainer{}
				for i := 0; i < 8; i++ {
					tc = append(tc, Metric{
						arch:               "rngd",
						core:               strconv.Itoa(i),
						device:             "npu0",
						kubernetesNodeName: "node",
						uuid:               "uuid",
						peUtilization:      float64(90),
					})
				}
				return tc
			}(),
			expected: `
# HELP furiosa_npu_core_utilization The current core utilization of NPU device
# TYPE furiosa_npu_core_utilization gauge
furiosa_npu_core_utilization{arch="rngd",core="0",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="1",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="2",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="3",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="4",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="5",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="6",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="7",device="npu0",kubernetes_node_name="node",uuid="uuid"} 90
`,
		},
	}

	cu := &coreUtilizationCollector{}
	cu.Register()
	for _, tc := range tests {
		err := cu.postProcess(tc.source)
		if err != nil {
			t.Errorf("unexpected error: %s\n", err)
		}

		err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(head+tc.expected), "furiosa_npu_core_utilization")
		if err != nil {
			t.Errorf("unexpected error: %s\n", err)
		}
	}
}

func TestCoreUtilizationCollector_Collect(t *testing.T) {
	//TODO: add test cases with mock device data
}
