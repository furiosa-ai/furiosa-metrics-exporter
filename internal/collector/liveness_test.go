package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

const head = `
# HELP furiosa_npu_alive The liveness of NPU device
# TYPE furiosa_npu_alive gauge
`

func TestLivenessCollector_PostProcessing(t *testing.T) {
	tests := []struct {
		description string
		source      MetricContainer
		expected    string
	}{
		{
			description: "liveness is true",
			source: func() MetricContainer {
				tc := MetricContainer{}
				metric := newMetric()
				metric[arch] = "rngd"
				metric[core] = "0-7"
				metric[device] = "npu0"
				metric[uuid] = "uuid"
				metric[liveness] = true
				tc = append(tc, metric)
				return tc
			}(),
			expected: `
furiosa_npu_alive{arch="rngd",container="",core="0-7",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="",pert_version="",pod="",uuid="uuid"} 1
`,
		},
		{
			description: "liveness is false",
			source: func() MetricContainer {
				tc := MetricContainer{}
				metric := newMetric()
				metric[arch] = "rngd"
				metric[core] = "0-7"
				metric[device] = "npu0"
				metric[uuid] = "uuid"
				metric[liveness] = false
				tc = append(tc, metric)
				return tc
			}(),
			expected: `
furiosa_npu_alive{arch="rngd",container="",core="0-7",device="npu0",driver_version="",firmware_version="",hostname="",namespace="",pci_bus_id="",pert_version="",pod="",uuid="uuid"} 0
`,
		},
	}

	p := &livenessCollector{}
	p.Register()
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			err := p.postProcess(tc.source)
			assert.NoError(t, err)

			err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(head+tc.expected), "furiosa_npu_alive")
			assert.NoError(t, err)
		})
	}
}

func TestLivenessCollector_Collect(t *testing.T) {
	//TODO: add testcases with device mock
}
