package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
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
			source: MetricContainer{
				{
					arch:     "rngd",
					core:     "0-7",
					device:   "npu0",
					uuid:     "uuid",
					liveness: true,
				},
			},
			expected: `
furiosa_npu_alive{arch="rngd",core="0-7",device="npu0",uuid="uuid"} 1
`,
		},
		{
			description: "liveness is false",
			source: MetricContainer{
				{
					arch:     "rngd",
					core:     "0-7",
					device:   "npu0",
					uuid:     "uuid",
					liveness: false,
				},
			},
			expected: `
furiosa_npu_alive{arch="rngd",core="0-7",device="npu0",uuid="uuid"} 0
`,
		},
	}

	p := &livenessCollector{}
	p.Register()
	for _, tc := range tests {
		err := p.postProcess(tc.source)
		if err != nil {
			t.Errorf("unexpected error:%s\n", err)
		}

		err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(head+tc.expected), "furiosa_npu_alive")
		if err != nil {
			t.Errorf("unexpected error:%s\n", err)
		}
	}
}

func TestLivenessCollector_Collect(t *testing.T) {
	//TODO: add testcases with device mock
}
