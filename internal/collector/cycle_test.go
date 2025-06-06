package collector

import (
	"strconv"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func newFakeCycleCollector() Collector {
	return &cycleCollector{
		devices:       nil,
		metricFactory: nil,
		kubeResMapper: NewFakeKubeResourcesMapper(),
	}
}

func TestCycleCollector_PostProcessing(t *testing.T) {
	tests := []struct {
		description string
		source      MetricContainer
		expected    string
	}{
		{
			description: "random cycle metrics",
			source: func() MetricContainer {
				mc := MetricContainer{}
				for i := 0; i < 8; i++ {
					metric := newMetric()
					metric[arch] = "rngd"
					metric[core] = strconv.Itoa(i)
					metric[device] = "npu0"
					metric[uuid] = "uuid"
					metric[bdf] = "bdf"
					metric[taskExecutionCycle] = float64(1234)
					metric[totalCycleCount] = float64(5678)

					mc = append(mc, metric)
				}

				return mc
			}(),
			expected: `
# HELP furiosa_npu_task_execution_cycle The current task execution cycle of NPU device
# TYPE furiosa_npu_task_execution_cycle counter
furiosa_npu_task_execution_cycle{arch="rngd",core="0",device="npu0",pci_bus_id="bdf",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",core="1",device="npu0",pci_bus_id="bdf",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",core="2",device="npu0",pci_bus_id="bdf",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",core="3",device="npu0",pci_bus_id="bdf",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",core="4",device="npu0",pci_bus_id="bdf",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",core="5",device="npu0",pci_bus_id="bdf",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",core="6",device="npu0",pci_bus_id="bdf",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",core="7",device="npu0",pci_bus_id="bdf",uuid="uuid"} 1234
# HELP furiosa_npu_total_cycle_count The current total cycle count of NPU device
# TYPE furiosa_npu_total_cycle_count counter
furiosa_npu_total_cycle_count{arch="rngd",core="0",device="npu0",pci_bus_id="bdf",uuid="uuid"} 5678
furiosa_npu_total_cycle_count{arch="rngd",core="1",device="npu0",pci_bus_id="bdf",uuid="uuid"} 5678
furiosa_npu_total_cycle_count{arch="rngd",core="2",device="npu0",pci_bus_id="bdf",uuid="uuid"} 5678
furiosa_npu_total_cycle_count{arch="rngd",core="3",device="npu0",pci_bus_id="bdf",uuid="uuid"} 5678
furiosa_npu_total_cycle_count{arch="rngd",core="4",device="npu0",pci_bus_id="bdf",uuid="uuid"} 5678
furiosa_npu_total_cycle_count{arch="rngd",core="5",device="npu0",pci_bus_id="bdf",uuid="uuid"} 5678
furiosa_npu_total_cycle_count{arch="rngd",core="6",device="npu0",pci_bus_id="bdf",uuid="uuid"} 5678
furiosa_npu_total_cycle_count{arch="rngd",core="7",device="npu0",pci_bus_id="bdf",uuid="uuid"} 5678
`,
		},
	}

	collector := newFakeCycleCollector()
	collector.Register()
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			err := collector.postProcess(tc.source)
			assert.Nil(t, err)

			err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(head+tc.expected), "furiosa_npu_task_execution_cycle")
			assert.NoError(t, err)
		})
	}
}

func TestCycleCollector_Collect(t *testing.T) {
	//TODO: add test cases with mock device data
}
