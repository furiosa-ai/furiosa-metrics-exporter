package collector

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func newFakeCoreUtilizationCollector() Collector {
	return &coreUtilizationCollector{
		devices:       nil,
		metricFactory: nil,
		kubeResMapper: NewFakeKubeResourcesMapper(),
	}
}

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
					metric := newMetric()
					metric[arch] = "rngd"
					metric[core] = strconv.Itoa(i)
					metric[device] = "npu0"
					metric[uuid] = "uuid"
					metric[bdf] = "bdf"
					metric[peUtilization] = float64(90)
					tc = append(tc, metric)
				}
				return tc
			}(),
			expected: `
# HELP furiosa_npu_core_utilization The current core utilization of NPU device
# TYPE furiosa_npu_core_utilization gauge
furiosa_npu_core_utilization{arch="rngd",core="0",device="npu0",pci_bus_id="bdf",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="1",device="npu0",pci_bus_id="bdf",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="2",device="npu0",pci_bus_id="bdf",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="3",device="npu0",pci_bus_id="bdf",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="4",device="npu0",pci_bus_id="bdf",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="5",device="npu0",pci_bus_id="bdf",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="6",device="npu0",pci_bus_id="bdf",uuid="uuid"} 90
furiosa_npu_core_utilization{arch="rngd",core="7",device="npu0",pci_bus_id="bdf",uuid="uuid"} 90

`,
		},
	}

	collector := newFakeCoreUtilizationCollector()
	collector.Register()

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			err := collector.postProcess(tc.source)
			assert.NoError(t, err)

			err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(head+tc.expected), "furiosa_npu_core_utilization")
			assert.NoError(t, err)
		})
	}
}

func TestCoreUtilizationCollector_Collect(t *testing.T) {
	//TODO: add test cases with mock device data
}
