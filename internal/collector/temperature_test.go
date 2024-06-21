package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestTempCollector_PostProcessing(t *testing.T) {
	c := &temperatureCollector{}
	c.Register()

	tc := MetricContainer{
		{
			arch:    "rngd",
			core:    "0-7",
			device:  "npu0",
			uuid:    uuid,
			ambient: float64(35),
			peak:    float64(39),
		},
	}

	err := c.postProcess(tc)
	if err != nil {
		t.Errorf("unexpected error:%s\n", err)
	}

	expected := `
# HELP furiosa_npu_hw_temperature The current temperatureCollector of NPU components
# TYPE furiosa_npu_hw_temperature gauge
furiosa_npu_hw_temperature{arch="rngd",core="0-7",device="npu0",label="peak",uuid="uuid"} 39
furiosa_npu_hw_temperature{arch="rngd",core="0-7",device="npu0",label="ambient",uuid="uuid"} 35
`
	err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "furiosa_npu_hw_temperature")
	if err != nil {
		t.Errorf("unexpected error:%s\n", err)
	}
}

func TestTempCollector_Collect(t *testing.T) {
	//TODO: add testcases with device mock
}
