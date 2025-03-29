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
			description: "random task execution cycle metrics",
			source: func() MetricContainer {
				tc := MetricContainer{}
				for i := 0; i < 8; i++ {
					metric := newMetric()
					metric[arch] = "rngd"
					metric[core] = strconv.Itoa(i)
					metric[device] = "npu0"
					metric[uuid] = "uuid"
					metric[bdf] = "bdf"
					metric[taskExecutionCycle] = float64(1234)
					tc = append(tc, metric)
				}
				return tc
			}(),
			expected: `
# HELP furiosa_npu_task_execution_cycle The current task execution cycle of NPU device
# TYPE furiosa_npu_task_execution_cycle counter
furiosa_npu_task_execution_cycle{arch="rngd",container="",core="0",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",container="",core="1",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",container="",core="2",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",container="",core="3",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",container="",core="4",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",container="",core="5",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",container="",core="6",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 1234
furiosa_npu_task_execution_cycle{arch="rngd",container="",core="7",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="bdf",pert_version="",pod="",uuid="uuid"} 1234
`,
		},
	}

	tec := &taskExecutionCycleCollector{}
	tec.Register()
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			err := tec.postProcess(tc.source)
			assert.NoError(t, err)

			err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(head+tc.expected), "furiosa_npu_task_execution_cycle")
			assert.NoError(t, err)
		})
	}
}

func TestPerformanceCounterCollector_Collect(t *testing.T) {
	//TODO: add test cases with mock device data
}
