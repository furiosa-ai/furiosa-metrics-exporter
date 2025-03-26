package collector

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCoreFrequencyCollector_PostProcessing(t *testing.T) {
	tests := []struct {
		description string
		source      MetricContainer
		expected    string
	}{
		{
			description: "random core frequency metrics",
			source: func() MetricContainer {
				tc := MetricContainer{}
				for i := 0; i < 8; i++ {
					tc = append(tc, Metric{
						arch:                "rngd",
						core:                strconv.Itoa(i),
						device:              "npu0",
						uuid:                "uuid",
						peFrequency:         uint32(2000),
						kubernetesNode:      "node",
						kubernetesNamespace: "namespace",
						kubernetesPod:       "pod",
						kubernetesContainer: "container",
					})
				}
				return tc
			}(),
			expected: `
# HELP furiosa_npu_core_frequency The current core frequency of NPU device (MHz)
# TYPE furiosa_npu_core_frequency gauge
furiosa_npu_core_frequency{arch="rngd",core="0",device="npu0",kubernetes_container_name="container",kubernetes_namespace_name="namespace",kubernetes_node_name="",kubernetes_pod_name="pod",uuid="uuid"} 2000
furiosa_npu_core_frequency{arch="rngd",core="1",device="npu0",kubernetes_container_name="container",kubernetes_namespace_name="namespace",kubernetes_node_name="",kubernetes_pod_name="pod",uuid="uuid"} 2000
furiosa_npu_core_frequency{arch="rngd",core="2",device="npu0",kubernetes_container_name="container",kubernetes_namespace_name="namespace",kubernetes_node_name="",kubernetes_pod_name="pod",uuid="uuid"} 2000
furiosa_npu_core_frequency{arch="rngd",core="3",device="npu0",kubernetes_container_name="container",kubernetes_namespace_name="namespace",kubernetes_node_name="",kubernetes_pod_name="pod",uuid="uuid"} 2000
furiosa_npu_core_frequency{arch="rngd",core="4",device="npu0",kubernetes_container_name="container",kubernetes_namespace_name="namespace",kubernetes_node_name="",kubernetes_pod_name="pod",uuid="uuid"} 2000
furiosa_npu_core_frequency{arch="rngd",core="5",device="npu0",kubernetes_container_name="container",kubernetes_namespace_name="namespace",kubernetes_node_name="",kubernetes_pod_name="pod",uuid="uuid"} 2000
furiosa_npu_core_frequency{arch="rngd",core="6",device="npu0",kubernetes_container_name="container",kubernetes_namespace_name="namespace",kubernetes_node_name="",kubernetes_pod_name="pod",uuid="uuid"} 2000
furiosa_npu_core_frequency{arch="rngd",core="7",device="npu0",kubernetes_container_name="container",kubernetes_namespace_name="namespace",kubernetes_node_name="",kubernetes_pod_name="pod",uuid="uuid"} 2000
`,
		},
	}

	cu := &coreFrequencyCollector{}
	cu.Register()
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			err := cu.postProcess(tc.source)
			assert.NoError(t, err)

			err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(head+tc.expected), "furiosa_npu_core_frequency")
			assert.NoError(t, err)
		})
	}
}

func TestCoreFrequencyCollector_Collect(t *testing.T) {
	//TODO: add test cases with mock device data
}
