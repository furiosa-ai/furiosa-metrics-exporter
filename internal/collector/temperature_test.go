package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestTempCollector_PostProcessing(t *testing.T) {
	c := &temperatureCollector{}
	c.Register()

	tc := MetricContainer{
		{
			arch:               "rngd",
			core:               "0-7",
			device:             "npu0",
			uuid:               uuid,
			kubernetesNodeName: "node",
			ambient:            float64(35),
			peak:               float64(39),
			pod:                "test",
		},
	}
	err := c.postProcess(tc)
	assert.NoError(t, err)

	expected := `
# HELP furiosa_npu_hw_temperature The current temperature of NPU device
# TYPE furiosa_npu_hw_temperature gauge
furiosa_npu_hw_temperature{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="peak",pod="test",uuid="uuid"} 39
furiosa_npu_hw_temperature{arch="rngd",core="0-7",device="npu0",kubernetes_node_name="node",label="ambient",pod="test",uuid="uuid"} 35
`
	err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "furiosa_npu_hw_temperature")
	assert.NoError(t, err)
}

func TestTempCollector_Collect(t *testing.T) {
	//TODO: add testcases with device mock
}
