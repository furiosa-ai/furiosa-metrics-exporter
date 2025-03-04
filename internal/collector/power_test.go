package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestPowerCollector_PostProcessing(t *testing.T) {
	p := &powerCollector{}
	p.Register()

	tc := MetricContainer{
		{
			arch:               "rngd",
			core:               "0-7",
			device:             "npu0",
			kubernetesNodeName: "node",
			rms:                float64(4795000),
			uuid:               "uuid",
			pod:                "test",
		},
	}
	err := p.postProcess(tc)
	assert.NoError(t, err)

	expected := `
# HELP furiosa_npu_hw_power The current power of NPU device
# TYPE furiosa_npu_hw_power gauge
furiosa_npu_hw_power{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="rms",pod="test",uuid="uuid"} 4795000
`

	err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "furiosa_npu_hw_power")
	assert.NoError(t, err)
}

func TestPowerCollector_Collect(t *testing.T) {
	//TODO: add testcases with device mock
}
