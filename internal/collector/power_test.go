package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPowerCollector_PostProcessing(t *testing.T) {
	p := &powerCollector{}
	p.Register()

	tc := MetricContainer{
		{
			arch:   "rngd",
			core:   "0-7",
			device: "npu0",
			uuid:   "uuid",
			rms:    float64(4795000),
		},
	}

	err := p.postProcess(tc)
	if err != nil {
		t.Errorf("unexpected error:%s\n", err)
	}

	expected := `
# HELP furiosa_npu_hw_power The current power of NPU components
# TYPE furiosa_npu_hw_power gauge
furiosa_npu_hw_power{arch="rngd",core="0-7",device="npu0",label="rms",uuid="uuid"} 4795000
`
	err = testutil.GatherAndCompare(prometheus.DefaultGatherer, strings.NewReader(expected), "furiosa_npu_hw_power")
	if err != nil {
		t.Errorf("unexpected error:%s\n", err)
	}
}

func TestPowerCollector_Collect(t *testing.T) {
	//TODO: add testcases with device mock
}
